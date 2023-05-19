package main

import (
	"regexp"
	"testing"
)

func TestCreateToken(t *testing.T) {
	jwt := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJkc3Nmc2RkZmRmIiwic2NvcGUiOlsiaWRlbnRpdHkiXSwiZXhwIjoxNjg5NzYyOTIwLCJpc3MiOiJodHRwczovL29hdXRoLmdyYXZpdGFsaWEuY29tIiwiaWF0IjoxNjg0NTAzMzIwfQ.pN7A1L_iO15hzopjY1IzI6x45sHszKULv1cMDPH4sPQ6hyUeCAQ59Fe_ix8RyiXDafRWDGt4MwQiUKUjiHiB9wzSd_H9-_FIFDaAUXmTQx2mSxDD8LsWGJHz8MHuVRBn02dZxocrU3vcywHgngFE2FPuhS2IT33LB85dsPQvXdofmRHggkt-QReLZTJ6lyeAQX8L1yjwQR58K4FY43OD9twIm0fMBcoKi5W_6hwC1B33OCyA7oE945B86QM7fNm_Gbh7lsittaZ9dckuPIN_FnUhpEp8SIJIm9bQMu3vjyeLJFPtfChexQ98FOUJvWwlb1jb7fM-BVWXyCFaHIKT-J55-v7XqYNDwpORvynvKrzz4400WQpV_0L712x_Z8dM7pS4mBZAI5nCE2x9DAPbh-UodAqg_llVPnS4OeOvUD0rsH8GFW1IE1KsdRlJFHrBd30DlFb2lgsOSllj6YK7aoA_MpK5GMOcLx0sel83C7qbPgFHbQr_yyJyc1VyzuV9voM8y9tscML6vzA0yBuVc3j9pvDtt6WL1_S-5lYdmr6zlX1xj0Nb_DOD-ObwuAIaucYXpB_uVm-A5OgYpFLw5cwyNAvQAP0Y_0UVQu30tff0_tVJ80BxqfkRi5jLs5t-9y5Dc4vRWTWCJy21IIZcZ3P6KCcsPw0U_JRK3ir5TAA"

	jwtCheck := regexp.MustCompile(`(^[A-Za-z0-9-_]*\.[A-Za-z0-9-_]*\.[A-Za-z0-9-_]*$)`)
	if !jwtCheck.MatchString(jwt) {
		t.Fatalf(`CreateToken("test") = %q, want match for %#q, nil`, jwt, jwtCheck)
	}
}
