package identity

import "testing"

func TestPasswordRoundTripAndRehash(t *testing.T) {
	parameters := DefaultPasswordParameters()
	encoded, err := HashPassword("correct horse battery staple", parameters)
	if err != nil {
		t.Fatal(err)
	}
	valid, rehash, err := VerifyPassword("correct horse battery staple", encoded, parameters)
	if err != nil || !valid || rehash {
		t.Fatalf("valid=%v rehash=%v err=%v", valid, rehash, err)
	}
	stronger := parameters
	stronger.Iterations++
	valid, rehash, err = VerifyPassword("correct horse battery staple", encoded, stronger)
	if err != nil || !valid || !rehash {
		t.Fatalf("valid=%v rehash=%v err=%v", valid, rehash, err)
	}
	valid, _, err = VerifyPassword("wrong password", encoded, parameters)
	if err != nil || valid {
		t.Fatalf("valid=%v err=%v", valid, err)
	}
}
