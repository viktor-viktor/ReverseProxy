package Authentication

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestAuthentication_Validate(t *testing.T) {
	var auth Authentication = Authentication{Name: "test", Auth_scheme: "http", Auth_type: "epp",
		Auth_addr: "adas", Url_path: "ADdas", Req_headers: []string {"Adsa", "sad"}}

	err := auth.Validate()
	if err != nil {
		t.Error(err)
	}

	auth.Name = ""
	err = auth.Validate()
	if err == nil {
		t.Error("Error should be returned when Name = ''")
	}

	auth.Name = "rest"
	auth.Auth_scheme = ""
	err = auth.Validate()
	if err == nil {
		t.Error("Error should be returned when Auth_scheme = ''")
	}
	auth.Auth_scheme = "dsafd"
	err = auth.Validate()
	if err == nil {
		t.Error("Supported scheme is only 'http'")
	}

	auth.Auth_scheme = "http"
	auth.Auth_addr = ""
	err = auth.Validate()
	if err == nil {
		t.Error("Error should be returned when auth_addr = ''")
	}

	auth.Auth_addr = "asd"
	auth.Auth_type = ""
	err = auth.Validate()
	if err == nil {
		t.Error("Error should be returned when auth_type = ''")
	}
	auth.Auth_type = "ads"
	err = auth.Validate()
	if err == nil {
		t.Error("Supported auth type only 'epp' (endpoint per permission")
	}

	auth.Auth_type = "epp"
	auth.Url_path = ""
	err = auth.Validate()
	if err == nil {
		t.Error("Error should be returned when Auth_url = ''")
	}

	auth.Url_path = "sgjdf"
	auth.Req_headers = nil
	err = auth.Validate()
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Auth.Validate test is finished !")
}

func TestRegisterMiddleware(t *testing.T) {
	var auth Authentication = Authentication{Auth_type: "epp"}
	_,err := RegisterMiddleware(auth)
	if err != nil {
		t.Error(err)
	}

	auth.Auth_type = "asd"
	_,err = RegisterMiddleware(auth)
	if err == nil {
		t.Error("Must not be able to register middleware with auth type other then 'epp'")
	}
}

func TestReadAuthFromFile(t *testing.T) {

	// valid data
	byt := []byte(`{"Auth": [{
		"name": "public",
		"auth_addr": "localhost:5000",
		"auth_type": "epp",
		"auth_scheme": "http",
		"url_path": "/api/authorization/public",
		"req_headers": [
			"Authorization"
	]
	}]}`)

	var mock map[string]interface{}
	json.Unmarshal(byt, &mock)

	_, err := ReadAuthFromFile(mock["Auth"])
	if err != nil {
		t.Error(err)
	}

	//invalid data
	byt = []byte(`{"Auth": {
		"name": "public",
		"auth_addr": "localhost:5000",
		"auth_type": "epp",
		"auth_scheme": "http",
		"url_path": "/api/authorization/public",
		"req_headers": [
			"Authorization"
	]
	}}`)
	json.Unmarshal(byt, &mock)
	_, err = ReadAuthFromFile(mock["auth"])
	if err == nil {
		t.Error("Shouldn't be able to cast array to ")
	}


}
