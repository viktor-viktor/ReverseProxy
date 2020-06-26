package Protocol

import "github.com/mitchellh/mapstructure"

var Protocols []Protocol = []Protocol{}

type Protocol struct {
	Type string
	Port int
	CertPath string
	KeyPath string
}

func (p *Protocol) Validate() {
	if p.Type != "http" && p.Type != "https" {
		// log error
		panic("Unsupported type protocol type is used: " + p.Type)
	}

	if p.Type == "https" && (p.CertPath == "" || p.KeyPath == "") {
		panic("https protocol should contain both CertPath and KeyPath variables")
	}
}

func ReadProtocolFormFile(prot interface{}) []Protocol {
	var rv = []Protocol{}

	prot2, ok := prot.([]interface{})
	if ok == false {
		// log error
		panic("Can't cast interface{} to []interface{}. Make sure you have an array under " +
			"'Protocols' category")
	}

	for _, v := range(prot2) {
		var tmp Protocol
		err := mapstructure.Decode(v, &tmp)
		if err != nil {panic("Can't decide protocol settings object. " + err.Error())}

		tmp.Validate()
		rv = append(rv, tmp)
	}

	return rv
}

func InitProtocols(file map[string]interface{}) {
	if v, exist := file["Protocols"]; exist {
		Protocols = ReadProtocolFormFile(v)
	} else {
		panic("Can't find 'Protocols' section in settings file")
	}
}
