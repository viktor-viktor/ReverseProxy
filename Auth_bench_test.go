package main

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"testing"
)

func BenchmarkReadAuthFromFile(b *testing.B) {
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

	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		_, err := ReadAuthFromFile(mock["Auth"])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func BenchmarkAuthentication_Validate(b *testing.B) {
	byt := []byte(`{
		"name": "public",
		"auth_addr": "localhost:5000",
		"auth_type": "epp",
		"auth_scheme": "http",
		"url_path": "/api/authorization/public",
		"req_headers": [
			"Authorization"
	]
	}`)

	var mock map[string]interface{}
	json.Unmarshal(byt, &mock)

	var auth Authentication
	mapstructure.Decode(mock, &auth)

	b.ResetTimer()

	for i:=0; i<b.N; i++ {
		err := auth.Validate()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func BenchmarkRegisterMiddleware(b *testing.B) {
	byt := []byte(`{
		"name": "public",
		"auth_addr": "localhost:5000",
		"auth_type": "epp",
		"auth_scheme": "http",
		"url_path": "/api/authorization/public",
		"req_headers": [
			"Authorization"
	]
	}`)

	var mock map[string]interface{}
	json.Unmarshal(byt, &mock)

	var auth Authentication
	mapstructure.Decode(mock, &auth)

	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		_, err := RegisterMiddleware(auth)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func BenchmarkDefaultAuthMiddleware(b *testing.B) {
	byt := []byte(`{
		"name": "public",
		"auth_addr": "localhost:5000",
		"auth_type": "epp",
		"auth_scheme": "http",
		"url_path": "/api/authorization/public",
		"req_headers": [
			"Authorization"
	]
	}`)

	var mock map[string]interface{}
	json.Unmarshal(byt, &mock)

	var auth Authentication
	mapstructure.Decode(mock, &auth)

	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		DefaultAuthMiddleware(auth)
	}
}