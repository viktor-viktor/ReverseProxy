package main

import (
	"testing"
)

func TestEndpointSettings_Validate(t *testing.T) {
	var end EndpointSettings = EndpointSettings{Entry_url: "ad", Redir_url: "Asd", Redir_addr: "adA",
		Methods: []string{"POST", "GET"}, Use_auth: true, Auth_name: "adasd"}

	err := end.Validate()
	if err != nil {t.Error(err)}

	end.Entry_url = ""
	err = end.Validate()
	if err == nil {t.Error("Validate() should fail if 'entry_url' is empty")}

	end.Entry_url = "sfdsdf"
	end.Redir_addr = ""
	err = end.Validate()
	if err == nil {t.Error("Validate should faild if 'redire_url' is empty")}

	end.Redir_addr = "adasd"
	end.Methods = append(end.Methods, "sdasda")
	err = end.Validate()
	if err == nil {t.Error("Validate should faild if 'methods' has invalid method inside")}

	end.Methods = end.Methods[:len(end.Methods) - 1]
	err = end.Validate()
	if err != nil { t.Error(err) }

	end.Redir_url = ""
	err = end.Validate()
	if err != nil {t.Error(err)}

	end.Use_auth = false
	err = end.Validate()
	if err != nil {t.Error(err)}

	end.Use_auth = true
	end.Auth_name = ""
	err = end.Validate()
	if err == nil {t.Error("Validation() should fail if auth_name is empty while use_auth is true")}

	end.Methods = []string{}
	end.Use_auth = false
	err = end.Validate()
	if err != nil {t.Error(err)}
}
