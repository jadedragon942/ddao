module github.com/jadedragon942/ddao/examples/ldapd

go 1.24.4

require (
	github.com/go-ldap/ldap/v3 v3.4.8
	github.com/jadedragon942/ddao v0.0.0
)

replace github.com/jadedragon942/ddao => ../..

require (
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.28 // indirect
	golang.org/x/crypto v0.37.0 // indirect
)
