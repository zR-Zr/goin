package auth

import (
	"testing"

	"github.com/zR-Zr/goin/pkg/config"
)

func TestJwt(t *testing.T) {
	cfg, err := config.LoadConfig("../../exapmle/config.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	auther := NewJWTAuth(cfg)

	u1 := JWTUser{
		ID:          123456,
		Username:    "admin",
		Type:        "admin",
		IsAnonymous: false,
	}

	accessToken, err := auther.GenrateToken(&u1)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("accessToken: ", accessToken)
	}
}

func TestParseToken(t *testing.T) {
	cfg, err := config.LoadConfig("../../exapmle/config.yaml")
	t.Log(cfg)
	if err != nil {
		t.Error(err)
	}
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MTIzNDU2LCJ1c2VybmFtZSI6ImFkbWluIiwidHlwZSI6ImFkbWluIiwiaXNzIjoiYWRtaW4iLCJzdWIiOiJhY2Nlc3MiLCJleHAiOjE3MzI3NzIwOTgsIm5iZiI6MTczMjc3MTE5OCwiaWF0IjoxNzMyNzcxMTk4LCJqdGkiOiJlNDNlMGU1OS02OTUyLTQ0NmMtODZiZC01MTYzYWMxYWZkNjEifQ.Cf9kt2lKzMaIEyFDcXRtaz46-EY0KoC8gx2Ig79vTzU"
	auther := NewJWTAuth(cfg)
	user, err := auther.ParseToken(token)
	if err != nil {
		t.Error(err)
	} else {
		// jsonUser, _ := json.Marshal(user)
		t.Log(user)
	}
}
