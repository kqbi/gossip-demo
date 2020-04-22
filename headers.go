package main

import "github.com/kqbi/gossip/base"

// Utility methods for creating headers.

func Via(e *endpoint, branch string) *base.ViaHeader {
	return &base.ViaHeader{
		&base.ViaHop{
			ProtocolName:    "SIP",
			ProtocolVersion: "2.0",
			Transport:       e.transport,
			Host:            e.host,
			Port:            &e.port,
			Params: base.NewParams().Add("branch", base.String{branch}),
			},
	}
}

func To(e *endpoint, tag string) *base.ToHeader {
	header := &base.ToHeader{
		DisplayName: base.String{e.displayName},
		Address: &base.SipUri{
			User: base.String{e.username},
			Host: e.host,
			UriParams: base.NewParams().Add("transport", base.String{e.transport}),
		},
		Params: base.NewParams(),
	}

	if tag != "" {
		header.Params.Add("tag", base.String{tag})
	}

	return header
}

func From(e *endpoint, tag string) *base.FromHeader {
	header := &base.FromHeader{
		DisplayName: base.String{e.displayName},
		Address: &base.SipUri{
			User: base.String{e.username},
			Host: e.host,
			UriParams: base.NewParams().Add("transport", base.String{e.transport}),
		},
		Params: base.NewParams(),
	}

	if tag != "" {
		header.Params.Add("tag", base.String{tag})
	}

	return header
}

func Contact(e *endpoint) *base.ContactHeader {
	return &base.ContactHeader{
		DisplayName: base.String{e.displayName},
		Address: &base.SipUri{
			User: base.String{e.username},
			Host: e.host,
			Port: &e.port,
		},
	}
}

func CSeq(seqno uint32, method base.Method) *base.CSeq {
	return &base.CSeq{
		SeqNo:      seqno,
		MethodName: method,
	}
}

func CallId(callid string) *base.CallId {
	header := base.CallId(callid)
	return &header
}

func ContentLength(l uint32) base.ContentLength {
	return base.ContentLength(l)
}
