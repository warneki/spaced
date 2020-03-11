package auth

import (
	"testing"
	"time"
)

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

func TestGenerateJWT(t *testing.T) {
	got := GenerateJWT("kitsune", "web")
	now := time.Now()

	iat := got.Issued.Time()
	oneMinAgo := now.Add(-1 * time.Minute)
	if !inTimeSpan(oneMinAgo, now, iat) {
		t.Errorf("Issued at is %s; want between %s and %s", iat.String(), oneMinAgo, now)
	}

	token, err := SignAndSerializeJWT(got)
	if err != nil {
		t.Error("Failed to sign and serialise the jwt: ", err)
	}

	_, err = VerifyJwt(token)

	if err != nil {
		t.Error("Failed to verify just created token: ", err)
	}
}
