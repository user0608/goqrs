package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"goqrs/envs"
	"io/fs"
	"os"
	"path"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ksaucedo002/answer/errores"
)

var (
	once       sync.Once
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)
var (
	jwtTokenLife time.Duration = 24 * time.Hour
)

func LoadRSAKeys() (err error) {
	jwtTokenLife, err = time.ParseDuration(envs.FindEnv("TOKEN_LIFE", "24h"))
	if err != nil {
		return err
	}
	keyPath := envs.FindEnv("GOQRS_RSA_PRIVATE", "certificates/id_rsa")
	pubPath := envs.FindEnv("GOQRS_RSA_PUBLIC", "certificates/id_rsa.pub")
	if keyPath == "" || pubPath == "" {
		keyPath = "jwt/rsa"
		pubPath = "jwt/rsa.pub"
	}
	if err := genRSAKeysIfNotExists(keyPath, pubPath); err != nil {
		return err
	}
	once.Do(func() {
		var private []byte
		var public []byte
		private, err = os.ReadFile(keyPath)
		if err != nil {
			return
		}
		public, err = os.ReadFile(pubPath)
		if err != nil {
			return
		}
		privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(private)
		if err != nil {
			return
		}
		publicKey, err = jwt.ParseRSAPublicKeyFromPEM(public)
		if err != nil {
			return
		}
	})
	return err
}
func genRSAKeysIfNotExists(privateKeyPath, pubKeyPath string) error {
	os.MkdirAll(path.Dir(privateKeyPath), fs.ModeDir|0755)
	os.MkdirAll(path.Dir(pubKeyPath), fs.ModeDir|0755)
	_, keyerr := os.Stat(privateKeyPath)
	_, puberr := os.Stat(pubKeyPath)
	if keyerr != nil || puberr != nil {
		if errors.Is(keyerr, os.ErrNotExist) || errors.Is(puberr, os.ErrNotExist) {
			os.Remove(privateKeyPath)
			os.Remove(pubKeyPath)
		} else {
			return fmt.Errorf("error: %v %v", keyerr, puberr)
		}
	}
	if keyerr == nil && puberr == nil {
		return nil
	}

	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return err
	}
	defer privateKeyFile.Close()

	publicKeyFile, err := os.Create(pubKeyPath)
	if err != nil {
		return err
	}
	defer publicKeyFile.Close()

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return err
	}
	if err := key.Validate(); err != nil {
		return fmt.Errorf("invalid: %w", err)
	}
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return fmt.Errorf("key-encode %w", err)
	}
	pub, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return fmt.Errorf("pkix %w", err)
	}
	publicKeyPEN := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pub,
	}
	//ParsePKIXPublicKey
	if err := pem.Encode(publicKeyFile, publicKeyPEN); err != nil {
		return fmt.Errorf("pub-encode %w", err)
	}
	return nil
}

type JWTValues struct {
	Username string `json:"username"`
}
type jwtCustomClaims struct {
	JWTValues
	jwt.RegisteredClaims
}

func GenToken(values JWTValues) (string, error) {
	now := time.Now()
	customClaims := jwtCustomClaims{
		JWTValues: values,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "ksaucedo",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtTokenLife)),
			Audience:  jwt.ClaimStrings{"qrsystems"},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, customClaims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", nil
	}
	return tokenString, nil

}
func ValidateToken(tokenString string) (values JWTValues, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtCustomClaims{}, verifyWithKey)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return JWTValues{}, errores.NewUnauthorizedf(nil, "[close] la sesión a expirado")
		}
		return JWTValues{}, errores.NewUnauthorizedf(nil, "[close] no se reconoció la sesión")
	}
	if !token.Valid {
		return JWTValues{}, errores.NewUnauthorizedf(nil, "[close] token invalido")
	}
	claims, ok := token.Claims.(*jwtCustomClaims)
	if !ok {
		return JWTValues{}, errores.NewInternalf(nil, "[close] claims invalidos")
	}
	return claims.JWTValues, nil
}
func verifyWithKey(token *jwt.Token) (any, error) {
	return publicKey, nil
}
