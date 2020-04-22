package main

import (
	"fmt"
	"time"

	"github.com/kqbi/gossip/base"
	"github.com/kqbi/gossip/log"
	"github.com/kqbi/gossip/transaction"
	"github.com/kqbi/gossip/transport"
)

type endpoint struct {
	// Sip Params
	displayName string
	username    string
	host        string
	port        uint16 // Listens on this port.

	// Transport Params
	proxy      string // Proxy host
	proxy_port uint16 // Proxy host
	transport  string // Sends using this transport. ("tcp" or "udp")

	// Internal guts
	tm       *transaction.Manager
	dialog   dialog
	dialogIx int
}

type dialog struct {
	callId    string
	to_tag    string // The tag in the To header.
	from_tag  string // The tag in the From header.
	currentTx txInfo // The current transaction.
	cseq      uint32
}

type txInfo struct {
	tx     transaction.Transaction // The underlying transaction.
	branch string                  // The via branch.
}

func (e *endpoint) Start() error {
	transport, err := transport.NewManager("udp")
	if err != nil {
		log.Severe("Failed to start transport manager %v\n", err)
	}
	tm, err := transaction.NewManager(transport, fmt.Sprintf("%v:%v", e.host, e.port))
	if err != nil {
		return err
	}

	e.tm = tm

	return nil
}

func (e *endpoint) ClearDialog() {
	caller.dialog = dialog{}
}

func (caller *endpoint) Invite(callee *endpoint) error {
	// Starting a dialog.
	callid := "thisisacall" + string(caller.dialogIx)
	tag := "tag." + caller.username + "." + caller.host
	branch := "z9hG4bK.callbranch.INVITE"
	caller.dialog.callId = callid
	caller.dialog.from_tag = tag
	caller.dialog.currentTx = txInfo{}
	caller.dialog.currentTx.branch = branch

	invite := base.NewRequest(
		base.INVITE,
		&base.SipUri{
			User: base.String{callee.username},
			Host: callee.host,
			Port: &callee.port,
		},
		"SIP/2.0",
		[]base.SipHeader{
			Via(caller, branch),
			To(callee, caller.dialog.to_tag),
			From(caller, caller.dialog.from_tag),
			Contact(caller),
			CSeq(caller.dialog.cseq, base.INVITE),
			CallId(callid),
			ContentLength(0),
		},
		"",
	)
	caller.dialog.cseq += 1

	log.Info("Sending: %v", invite.Short())
	var dest string
	var port uint16
	if len(callee.proxy) != 0 {
		dest = callee.proxy
		port = callee.proxy_port
	} else {
		dest = callee.host
		port = callee.port
	}
	tx := caller.tm.Send(invite, fmt.Sprintf("%v:%v", dest, port))
	caller.dialog.currentTx.tx = transaction.Transaction(tx)
	for {
		select {
		case r := <-tx.Responses():
			log.Info("Received response: %v", r.Short())
			log.Debug("Full form:\n%v\n", r.String())
			// Get To tag if present.
			tag, ok := r.Headers("To").(*base.ToHeader).Params.Get("tag")
			if ok {
				caller.dialog.to_tag = tag.(base.String).String()
			}

			switch {
			case r.StatusCode >= 300:
				// Call setup failed.
				return fmt.Errorf("callee sent negative response code %v.", r.StatusCode)
			case r.StatusCode >= 200:
				// Ack 200s manually.
				log.Info("Sending Ack")
				tx.Ack()
				return nil
			}
		case e := <-tx.Errors():
			log.Warn(e.Error())
			return e
		}
	}
}

func (caller *endpoint) Register(callee *endpoint) error {
	return caller.nonInvite(callee, base.REGISTER)
}

func (caller *endpoint) Bye(callee *endpoint) error {
	return caller.nonInvite(callee, base.BYE)
}

func (caller *endpoint) nonInvite(callee *endpoint, method base.Method) error {
	var request *base.Request
	caller.dialog.currentTx.branch = fmt.Sprintf("z9hG4bK.%v.%v", method, caller.dialog.cseq)
	if method != base.REGISTER {
		request = base.NewRequest(
			method,
			&base.SipUri{
				User: base.String{callee.username},
				Host: callee.host,
				Port: &callee.port,
			},
			"SIP/2.0",
			[]base.SipHeader{
				Via(caller, caller.dialog.currentTx.branch),
				To(callee, caller.dialog.to_tag),
				From(caller, caller.dialog.from_tag),
				Contact(caller),
				CSeq(caller.dialog.cseq, method),
				CallId(caller.dialog.callId),
				ContentLength(0),
			},
			"",
		)
	} else {
		request = base.NewRequest(
			method,
			&base.SipUri{
				Host: callee.host,
			},
			"SIP/2.0",
			[]base.SipHeader{
				Via(caller, caller.dialog.currentTx.branch),
				&base.ToHeader{
					DisplayName: base.String{caller.displayName},
					Address: &base.SipUri{
						User: base.String{caller.username},
						Host: callee.host,
					},
					Params: base.NewParams(),
				},
				&base.FromHeader{
					DisplayName: base.String{caller.displayName},
					Address: &base.SipUri{
						User: base.String{caller.username},
						Host: callee.host,
					},
					Params: base.NewParams(),
				},
				Contact(caller),
				CSeq(caller.dialog.cseq, method),
				CallId(caller.dialog.callId),
				&base.GenericHeader{"Expires", "3600"},
				ContentLength(0),
			},
			"",
		)
	}
	caller.dialog.cseq += 1

	log.Info("Sending: %v", request.Short())
	fmt.Printf("request.Short() :%s\n", request.String())
	var dest string
	var port uint16
	if len(caller.proxy) != 0 {
		dest = caller.proxy
		port = caller.proxy_port
	} else {
		dest = callee.host
		port = callee.port
	}
	tx := caller.tm.Send(request, fmt.Sprintf("%v:%v", dest, port))
	caller.dialog.currentTx.tx = transaction.Transaction(tx)
	for {
		select {
		case r := <-tx.Responses():
			log.Info("Received response: %v", r.Short())
			log.Debug("Full form:\n%v\n", r.String())
			switch {
			case r.StatusCode >= 300:
				// Failure (or redirect).
				return fmt.Errorf("callee sent negative response code %v.", r.StatusCode)
			case r.StatusCode >= 200:
				// Success.
				log.Info("Successful transaction")
				return nil
			}
		case e := <-tx.Errors():
			log.Warn(e.Error())
			return e
		}
	}
}

// Server side function.

func (e *endpoint) ServeInvite() {
	log.Info("Listening for incoming requests...")
	tx := <-e.tm.Requests()
	r := tx.Origin()
	log.Info("Received request: %v", r.Short())
	log.Debug("Full form:\n%v\n", r.String())

	e.dialog.callId = string(*r.Headers("Call-Id").(*base.CallId))

	// Send a 200 OK
	resp := base.NewResponse(
		"SIP/2.0",
		200,
		"OK",
		[]base.SipHeader{},
		"",
	)

	base.CopyHeaders("Via", tx.Origin(), resp)
	base.CopyHeaders("From", tx.Origin(), resp)
	base.CopyHeaders("To", tx.Origin(), resp)
	base.CopyHeaders("Call-Id", tx.Origin(), resp)
	base.CopyHeaders("CSeq", tx.Origin(), resp)
	resp.AddHeader(
		&base.ContactHeader{
			DisplayName: base.String{e.displayName},
			Address: &base.SipUri{
				User: base.String{e.username},
				Host: e.host,
			},
		},
	)

	log.Info("Sending 200 OK")
	<-time.After(1 * time.Second)
	tx.Respond(resp)

	ack := <-tx.Ack()

	log.Info("Received ACK")
	log.Debug("Full form:\n%v\n", ack.String())
}

func (e *endpoint) ServeNonInvite() {
	log.Info("Listening for incoming requests...")
	for tx := range e.tm.Requests() {
		go e.HandleNonInvite(tx)
	}
}

func (e *endpoint) HandleNonInvite(tx *transaction.ServerTransaction) {
	r := tx.Origin()
	log.Info("Received request: %v", r.Short())
	log.Debug("Full form:\n%v\n", r.String())

	// Send a 200 OK
	resp := base.NewResponse(
		"SIP/2.0",
		200,
		"OK",
		[]base.SipHeader{},
		"",
	)

	base.CopyHeaders("Via", tx.Origin(), resp)
	base.CopyHeaders("From", tx.Origin(), resp)
	base.CopyHeaders("To", tx.Origin(), resp)
	base.CopyHeaders("Call-Id", tx.Origin(), resp)
	base.CopyHeaders("CSeq", tx.Origin(), resp)
	if tx.Origin().Method == base.REGISTER {
		to := tx.Origin().Headers("To").(*base.ToHeader)
		resp.AddHeader(
			&base.ContactHeader{
				Address: &base.SipUri{
					User: to.Address.(*base.SipUri).User,
					Host: to.Address.(*base.SipUri).Host,
				},
			},
		)
	} else {
		resp.AddHeader(
			&base.ContactHeader{
				DisplayName: base.String{e.displayName},
				Address: &base.SipUri{
					User: base.String{e.username},
					Host: e.host,
				},
			},
		)
	}

	log.Info("Sending 200 OK")
	<-time.After(1 * time.Second)
	tx.Respond(resp)
}
