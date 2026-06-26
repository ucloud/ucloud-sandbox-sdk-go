package sandbox

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

func getSignature(path, operation, user, accessToken string, expirationSec *int) (sig string, exp *int64, err error) {
	if accessToken == "" {
		return "", nil, fmt.Errorf("access token is not set")
	}
	var expiration *int64
	if expirationSec != nil {
		v := time.Now().Unix() + int64(*expirationSec)
		expiration = &v
	}
	raw := fmt.Sprintf("%s:%s:%s:%s", path, operation, user, accessToken)
	if expiration != nil {
		raw = fmt.Sprintf("%s:%d", raw, *expiration)
	}
	hash := sha256.Sum256([]byte(raw))
	encoded := strings.TrimRight(base64.StdEncoding.EncodeToString(hash[:]), "=")
	return fmt.Sprintf("v1_%s", encoded), expiration, nil
}
