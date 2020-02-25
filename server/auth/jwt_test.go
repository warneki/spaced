package auth

import (
    "testing"
    "time"
)

func inTimeSpan(start, end, check time.Time) bool {
    return check.After(start) && check.Before(end)
}

func TestGenerateJWT (t *testing.T) {
    got := GenerateJWT("kitsune", []string{"web"})
    now := time.Now()

    iat := got.Claims().Get("iat").(time.Time)
    oneMinAgo := now.Add(-1 * time.Minute)
    if !inTimeSpan(oneMinAgo, now, iat ) {
       t.Errorf("Issued at is %s; want between %s and %s", iat.String(), oneMinAgo, now)
    }
}

