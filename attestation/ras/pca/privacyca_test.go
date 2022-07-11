package pca

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"testing"

	"github.com/google/go-tpm-tools/simulator"
	"github.com/google/go-tpm/tpm2"
)

var (
	cmdEnc        = "enc"
	decFlag       = "-d"
	encFlag       = "-e"
	aes128cbcFlag = "-aes-128-cbc"
	aes192cbcFlag = "-aes-192-cbc"
	aes256cbcFlag = "-aes-256-cbc"
	aes128cfbFlag = "-aes-128-cfb"
	aes192cfbFlag = "-aes-192-cfb"
	aes256cfbFlag = "-aes-256-cfb"
	aes128ofbFlag = "-aes-128-ofb"
	aes192ofbFlag = "-aes-192-ofb"
	aes256ofbFlag = "-aes-256-ofb"
	aes128ctrFlag = "-aes-128-ctr"
	aes192ctrFlag = "-aes-192-ctr"
	aes256ctrFlag = "-aes-256-ctr"
	kFlag         = "-K"
	ivFlag        = "-iv"
	inFlag        = "-in"
	outFlag       = "-out"
	decFile       = "./test.txt.dec"
	encFile       = "./test.txt.enc"
	decPKeyFile   = "./test-pkey.txt.dec"
	encPKeyFile   = "./test-pkey.txt.enc"
	textFile      = "./test.txt"
	base64Flag    = "-base64"
	plainText     = "Hello, world!"
	ivValue       = "1234567890abcdef"
	key16Value    = "1234567890abcdef"
	key24Value    = "123456789012345678901234"
	key32Value    = "12345678901234567890123456789012"

	cmdRsa        = "rsa"
	cmdPKey       = "pkey"
	cmdRsautl     = "rsautl"
	cmdPKeyutl    = "pkeyutl"
	cmdGenrsa     = "genrsa"
	cmdGenpkey    = "genpkey"
	rsaKeyFile    = "./rsaKey.pem"
	pKeyFile      = "./pKey.pem"
	rsaPubKeyFile = "./rsaPubKey.pem"
	pubKeyFile    = "./pubKey.pem"
	pubInFlag     = "-pubin"
	pubOutFlag    = "-pubout" // PKCS#8
	//pubOutFlag = "-RSAPublicKey_out"	// PKCS#1
	algorithmFlag = "-algorithm"
	rsaFlag       = "RSA"
	encryptFlag   = "-encrypt"
	decryptFlag   = "-decrypt"
	inKeyFlag     = "-inkey"
	pKeyOptFlag   = "-pkeyopt"
	rsa2048Flag   = "rsa_keygen_bits:2048"
	rsaOAEPFlag   = "rsa_padding_mode:oaep"
	rsaSHA256Flag = "rsa_oaep_md:sha256"
)

const (
	constRESULT     = "result error: decrypt(%d bytes)='%s', want(%d bytes)='%s'"
	constINVOKE     = "invoke %s error: %v"
	constCMD        = "params: %v, err output: %v"
	constPRIVATEERR = "read private key error"
	constPUBLICERR  = "read public key error"
	constRSAPRIVATE = "RSA PRIVATE KEY"
	constPRIVATE    = "PRIVATE KEY"
	constPUBLIC     = "PUBLIC KEY"
	constRD         = "ReadFile()=>"
	constWT         = "WriteFile()=>"
	constSE         = "SymmetricEncrypt()=>"
	constSD         = "SymmetricDecrypt()"
	constAE         = "AsymmetricEncrypt()=>"
	constAD         = "AsymmetricDecrypt()"
	constPPK        = "ParsePKIXPublicKey()"
	constP1K        = "ParsePKCS1PrivateKey()"
	constP8K        = "ParsePKCS8PrivateKey()"
)

func runCmd(t *testing.T, params []string) {
	cmd := exec.Command("openssl", params...)
	out, err := cmd.Output()
	if err != nil {
		t.Errorf(constCMD, params, out)
	}
}

func TestSymEnc(t *testing.T) {
	var testCases = []struct {
		params []string
		text   string
		key    string
		iv     string
		alg    uint16
		mod    uint16
	}{
		{[]string{cmdEnc, decFlag, aes128cbcFlag, inFlag, decFile, outFlag, textFile, base64Flag}, plainText, key16Value, ivValue, AlgAES, AlgCBC},
		{[]string{cmdEnc, decFlag, aes192cbcFlag, inFlag, decFile, outFlag, textFile, base64Flag}, plainText, key24Value, ivValue, AlgAES, AlgCBC},
		{[]string{cmdEnc, decFlag, aes256cbcFlag, inFlag, decFile, outFlag, textFile, base64Flag}, plainText, key32Value, ivValue, AlgAES, AlgCBC},
		{[]string{cmdEnc, decFlag, aes128cfbFlag, inFlag, decFile, outFlag, textFile, base64Flag}, plainText, key16Value, ivValue, AlgAES, AlgCFB},
		{[]string{cmdEnc, decFlag, aes192cfbFlag, inFlag, decFile, outFlag, textFile, base64Flag}, plainText, key24Value, ivValue, AlgAES, AlgCFB},
		{[]string{cmdEnc, decFlag, aes256cfbFlag, inFlag, decFile, outFlag, textFile, base64Flag}, plainText, key32Value, ivValue, AlgAES, AlgCFB},
		{[]string{cmdEnc, decFlag, aes128ofbFlag, inFlag, decFile, outFlag, textFile, base64Flag}, plainText, key16Value, ivValue, AlgAES, AlgOFB},
		{[]string{cmdEnc, decFlag, aes192ofbFlag, inFlag, decFile, outFlag, textFile, base64Flag}, plainText, key24Value, ivValue, AlgAES, AlgOFB},
		{[]string{cmdEnc, decFlag, aes256ofbFlag, inFlag, decFile, outFlag, textFile, base64Flag}, plainText, key32Value, ivValue, AlgAES, AlgOFB},
		{[]string{cmdEnc, decFlag, aes128ctrFlag, inFlag, decFile, outFlag, textFile, base64Flag}, plainText, key16Value, ivValue, AlgAES, AlgCTR},
		{[]string{cmdEnc, decFlag, aes192ctrFlag, inFlag, decFile, outFlag, textFile, base64Flag}, plainText, key24Value, ivValue, AlgAES, AlgCTR},
		{[]string{cmdEnc, decFlag, aes256ctrFlag, inFlag, decFile, outFlag, textFile, base64Flag}, plainText, key32Value, ivValue, AlgAES, AlgCTR},
	}
	defer func() {
		os.Remove(textFile)
		os.Remove(decFile)
	}()
	for _, tc := range testCases {
		ciphertext, err := SymmetricEncrypt(tc.alg, tc.mod, []byte(tc.key), []byte(tc.iv), []byte(tc.text))
		if err != nil {
			t.Errorf(constINVOKE, constSE+tc.text, err)
		}
		// must have the last character "\n", otherwise can't be decrypted by openssl.
		base64text := base64.StdEncoding.EncodeToString(ciphertext) + "\n"
		err = ioutil.WriteFile(decFile, []byte(base64text), 0644)
		if err != nil {
			t.Errorf(constINVOKE, constWT+decFile, err)
		}
		params := append(tc.params, kFlag)
		params = append(params, hex.EncodeToString([]byte(tc.key)))
		params = append(params, ivFlag)
		params = append(params, hex.EncodeToString([]byte(tc.iv)))
		runCmd(t, params)
		plaintext, err := ioutil.ReadFile(textFile)
		if err != nil {
			t.Errorf(constINVOKE, constRD+textFile, err)
		}
		if string(plaintext) != tc.text {
			t.Errorf(constRESULT, len(plaintext), string(plaintext), len(tc.text), tc.text)
		}
	}
}

func TestSymDec(t *testing.T) {
	var testCases = []struct {
		params []string
		text   string
		key    string
		iv     string
		alg    uint16
		mod    uint16
	}{
		{[]string{cmdEnc, encFlag, aes128cbcFlag, inFlag, textFile, outFlag, encFile, base64Flag}, plainText, key16Value, ivValue, AlgAES, AlgCBC},
		{[]string{cmdEnc, encFlag, aes192cbcFlag, inFlag, textFile, outFlag, encFile, base64Flag}, plainText, key24Value, ivValue, AlgAES, AlgCBC},
		{[]string{cmdEnc, encFlag, aes256cbcFlag, inFlag, textFile, outFlag, encFile, base64Flag}, plainText, key32Value, ivValue, AlgAES, AlgCBC},
		{[]string{cmdEnc, encFlag, aes128cfbFlag, inFlag, textFile, outFlag, encFile, base64Flag}, plainText, key16Value, ivValue, AlgAES, AlgCFB},
		{[]string{cmdEnc, encFlag, aes192cfbFlag, inFlag, textFile, outFlag, encFile, base64Flag}, plainText, key24Value, ivValue, AlgAES, AlgCFB},
		{[]string{cmdEnc, encFlag, aes256cfbFlag, inFlag, textFile, outFlag, encFile, base64Flag}, plainText, key32Value, ivValue, AlgAES, AlgCFB},
		{[]string{cmdEnc, encFlag, aes128ofbFlag, inFlag, textFile, outFlag, encFile, base64Flag}, plainText, key16Value, ivValue, AlgAES, AlgOFB},
		{[]string{cmdEnc, encFlag, aes192ofbFlag, inFlag, textFile, outFlag, encFile, base64Flag}, plainText, key24Value, ivValue, AlgAES, AlgOFB},
		{[]string{cmdEnc, encFlag, aes256ofbFlag, inFlag, textFile, outFlag, encFile, base64Flag}, plainText, key32Value, ivValue, AlgAES, AlgOFB},
		{[]string{cmdEnc, encFlag, aes128ctrFlag, inFlag, textFile, outFlag, encFile, base64Flag}, plainText, key16Value, ivValue, AlgAES, AlgCTR},
		{[]string{cmdEnc, encFlag, aes192ctrFlag, inFlag, textFile, outFlag, encFile, base64Flag}, plainText, key24Value, ivValue, AlgAES, AlgCTR},
		{[]string{cmdEnc, encFlag, aes256ctrFlag, inFlag, textFile, outFlag, encFile, base64Flag}, plainText, key32Value, ivValue, AlgAES, AlgCTR},
	}
	defer func() {
		os.Remove(textFile)
		os.Remove(encFile)
	}()
	for _, tc := range testCases {
		err := ioutil.WriteFile(textFile, []byte(tc.text), 0644)
		if err != nil {
			t.Errorf(constINVOKE, constWT+textFile, err)
		}
		params := append(tc.params, kFlag)
		params = append(params, hex.EncodeToString([]byte(tc.key)))
		params = append(params, ivFlag)
		params = append(params, hex.EncodeToString([]byte(tc.iv)))
		runCmd(t, params)
		base64text, _ := ioutil.ReadFile(encFile)
		ciphertext, _ := base64.StdEncoding.DecodeString(string(base64text))
		plaintext, err := SymmetricDecrypt(tc.alg, tc.mod, []byte(tc.key), []byte(tc.iv), ciphertext)
		if err != nil {
			t.Errorf(constINVOKE, constSD, err)
		}
		if string(plaintext) != tc.text {
			t.Errorf(constRESULT, len(plaintext), string(plaintext), len(tc.text), tc.text)
		}
	}
}

func genRsaKeys(t *testing.T) {
	genRsaKeyParams := []string{
		cmdGenrsa,
		outFlag,
		rsaKeyFile,
	}
	genRsaPubKeyParams := []string{
		cmdRsa,
		inFlag,
		rsaKeyFile,
		pubOutFlag,
		outFlag,
		rsaPubKeyFile,
	}
	runCmd(t, genRsaKeyParams)
	runCmd(t, genRsaPubKeyParams)
}

func rsaEncrypt(t *testing.T) {
	genRsaEncryptParams := []string{
		cmdRsautl,
		encryptFlag,
		inFlag,
		textFile,
		pubInFlag,
		inKeyFlag,
		rsaPubKeyFile,
		outFlag,
		encFile,
	}
	runCmd(t, genRsaEncryptParams)
}

func rsaDecrypt(t *testing.T) {
	genRsaDecryptParams := []string{
		cmdRsautl,
		decryptFlag,
		inFlag,
		encFile,
		inKeyFlag,
		rsaKeyFile,
		outFlag,
		decFile,
	}
	runCmd(t, genRsaDecryptParams)
}

func genPKeys(t *testing.T) {
	genPKeyParams := []string{
		cmdGenpkey,
		algorithmFlag,
		rsaFlag,
		pKeyOptFlag,
		rsa2048Flag,
		outFlag,
		pKeyFile,
	}
	genPubKeyParams := []string{
		cmdPKey,
		inFlag,
		pKeyFile,
		pubOutFlag,
		outFlag,
		pubKeyFile,
	}
	runCmd(t, genPKeyParams)
	runCmd(t, genPubKeyParams)
}

func pKeyEncPKCS1(t *testing.T) {
	genRsaEncryptParams := []string{
		cmdPKeyutl,
		encryptFlag,
		inFlag,
		textFile,
		pubInFlag,
		inKeyFlag,
		pubKeyFile,
		outFlag,
		encPKeyFile,
	}
	runCmd(t, genRsaEncryptParams)
}

func pKeyDecPKCS1(t *testing.T) {
	genRsaDecryptParams := []string{
		cmdPKeyutl,
		decryptFlag,
		inFlag,
		encPKeyFile,
		inKeyFlag,
		pKeyFile,
		outFlag,
		decPKeyFile,
	}
	runCmd(t, genRsaDecryptParams)
}

func pKeyEncOAEP(t *testing.T) {
	genRsaEncryptParams := []string{
		cmdPKeyutl,
		encryptFlag,
		inFlag,
		textFile,
		pubInFlag,
		inKeyFlag,
		pubKeyFile,
		pKeyOptFlag,
		rsaOAEPFlag,
		pKeyOptFlag,
		rsaSHA256Flag,
		outFlag,
		encPKeyFile,
	}
	runCmd(t, genRsaEncryptParams)
}

func pKeyDecOAEP(t *testing.T) {
	genRsaDecryptParams := []string{
		cmdPKeyutl,
		decryptFlag,
		inFlag,
		encPKeyFile,
		inKeyFlag,
		pKeyFile,
		pKeyOptFlag,
		rsaOAEPFlag,
		pKeyOptFlag,
		rsaSHA256Flag,
		outFlag,
		decPKeyFile,
	}
	runCmd(t, genRsaDecryptParams)
}

func delRsaKeys() {
	os.Remove(rsaKeyFile)
	os.Remove(rsaPubKeyFile)
	os.Remove(pKeyFile)
	os.Remove(pubKeyFile)
	os.Remove(textFile)
	os.Remove(encFile)
	os.Remove(decFile)
	os.Remove(encPKeyFile)
	os.Remove(decPKeyFile)
}

func testAsymEncSchemeNull(t *testing.T, alg, mod uint16, text string) {
	rsaPubKeyPEM, _ := ioutil.ReadFile(rsaPubKeyFile)
	keyBlock, _ := pem.Decode(rsaPubKeyPEM)
	if keyBlock == nil || keyBlock.Type != constPUBLIC {
		t.Errorf(constPUBLICERR)
	}
	rsaPubKey, err := x509.ParsePKIXPublicKey(keyBlock.Bytes)
	if err != nil {
		t.Errorf(constINVOKE, constPPK, err)
	}
	ciphertext, err := AsymmetricEncrypt(alg, mod, rsaPubKey, []byte(text), nil)
	if err != nil {
		t.Errorf(constINVOKE, constAE+text, err)
	}
	err = ioutil.WriteFile(encFile, ciphertext, 0644)
	if err != nil {
		t.Errorf(constINVOKE, constWT+encFile, err)
	}
	rsaDecrypt(t)
	plaintext, _ := ioutil.ReadFile(decFile)
	if string(plaintext) != text {
		t.Errorf(constRESULT, len(plaintext), string(plaintext), len(text), text)
	}
}

func testAsymEncSchemeAll(t *testing.T, alg, mod uint16, text string) {
	pubKeyPEM, _ := ioutil.ReadFile(pubKeyFile)
	pKeyBlock, _ := pem.Decode(pubKeyPEM)
	if pKeyBlock == nil || pKeyBlock.Type != constPUBLIC {
		t.Errorf(constPUBLICERR)
	}
	pubKey, err := x509.ParsePKIXPublicKey(pKeyBlock.Bytes)
	if err != nil {
		t.Errorf(constINVOKE, constPPK, err)
	}
	ciphertext2, err := AsymmetricEncrypt(alg, mod, pubKey, []byte(text), nil)
	if err != nil {
		t.Errorf(constINVOKE, constAE+text, err)
	}
	err = ioutil.WriteFile(encPKeyFile, ciphertext2, 0644)
	if err != nil {
		t.Errorf(constINVOKE, constWT+encPKeyFile, err)
	}
	if mod == AlgOAEP {
		pKeyDecOAEP(t)
	} else {
		pKeyDecPKCS1(t)
	}
	plaintext2, _ := ioutil.ReadFile(decPKeyFile)
	if string(plaintext2) != text {
		t.Errorf(constRESULT, len(plaintext2), string(plaintext2), len(text), text)
	}
}

func TestAsymEnc(t *testing.T) {
	var testCases = []struct {
		text string
		alg  uint16
		mod  uint16
	}{
		{plainText, AlgRSA, AlgNull},
		{plainText, AlgRSA, AlgOAEP},
	}
	defer delRsaKeys()
	genRsaKeys(t)
	genPKeys(t)
	for _, tc := range testCases {
		if tc.mod == AlgNull {
			testAsymEncSchemeNull(t, tc.alg, tc.mod, tc.text)
		}
		testAsymEncSchemeAll(t, tc.alg, tc.mod, tc.text)
	}
}

func testAsymDecSchemeNull(t *testing.T, alg, mod uint16, text string) {
	rsaEncrypt(t)
	rsaKeyPEM, _ := ioutil.ReadFile(rsaKeyFile)
	keyBlock, _ := pem.Decode(rsaKeyPEM)
	if keyBlock == nil || keyBlock.Type != constRSAPRIVATE {
		t.Errorf(constPRIVATEERR)
	}
	// rsa for PKCS#1
	rsaPriKey, err1 := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err1 != nil {
		t.Errorf(constINVOKE, constP1K, err1)
	}
	ciphertext, _ := ioutil.ReadFile(encFile)
	plaintext, err1 := AsymmetricDecrypt(alg, mod, rsaPriKey, ciphertext, nil)
	if err1 != nil {
		t.Errorf(constINVOKE, constAD, err1)
	}
	if string(plaintext) != text {
		t.Errorf(constRESULT, len(plaintext), string(plaintext), len(text), text)
	}
}

func testAsymDecSchemeAll(t *testing.T, alg, mod uint16, text string) {
	if mod == AlgOAEP {
		pKeyEncOAEP(t)
	} else {
		pKeyEncPKCS1(t)
	}
	pKeyPEM, _ := ioutil.ReadFile(pKeyFile)
	pKeyBlock, _ := pem.Decode(pKeyPEM)
	if pKeyBlock == nil || pKeyBlock.Type != constPRIVATE {
		t.Errorf(constPRIVATEERR)
	}
	// pkey for PKCS#8
	priKey, err := x509.ParsePKCS8PrivateKey(pKeyBlock.Bytes)
	if err != nil {
		t.Errorf(constINVOKE, constP8K, err)
	}
	ciphertext2, _ := ioutil.ReadFile(encPKeyFile)
	plaintext2, err := AsymmetricDecrypt(alg, mod, priKey, ciphertext2, nil)
	if err != nil {
		t.Errorf(constINVOKE, constAD, err)
	}
	if string(plaintext2) != text {
		t.Errorf(constRESULT, len(plaintext2), string(plaintext2), len(text), text)
	}
}

func TestAsymDec(t *testing.T) {
	var testCases = []struct {
		text string
		alg  uint16
		mod  uint16
	}{
		{plainText, AlgRSA, AlgNull},
		{plainText, AlgRSA, AlgOAEP},
	}
	defer delRsaKeys()
	genRsaKeys(t)
	genPKeys(t)
	for _, tc := range testCases {
		err := ioutil.WriteFile(textFile, []byte(tc.text), 0644)
		if err != nil {
			t.Errorf(constINVOKE, constWT+textFile, err)
		}

		if tc.mod == AlgNull {
			testAsymDecSchemeNull(t, tc.alg, tc.mod, tc.text)
		}
		testAsymDecSchemeAll(t, tc.alg, tc.mod, tc.text)
	}
}

func TestKDFa(t *testing.T) {
	var testCases = []struct {
		key      string
		label    string
		contextU string
		contextV string
		size     int
	}{
		{"123", "abc", "defad", "mmmm", 29},
	}
	for _, tc := range testCases {
		a, _ := KDFa(crypto.SHA256, []byte(tc.key), tc.label, []byte(tc.contextU), []byte(tc.contextV), tc.size)
		b, _ := tpm2.KDFa(tpm2.AlgSHA256, []byte(tc.key), tc.label, []byte(tc.contextU), []byte(tc.contextV), tc.size)
		if !bytes.Equal(a, b) {
			t.Errorf("KDFa can't match, %v, %v\n", a, b)
		}
	}
}

var (
	simulatorMutex sync.Mutex
)

func pubKeyToTPMPublic(ekPubKey crypto.PublicKey) *tpm2.Public {
	pub := DefaultKeyParams
	pub.RSAParameters.KeyBits = uint16(uint32(ekPubKey.(*rsa.PublicKey).N.BitLen()))
	pub.RSAParameters.ExponentRaw = uint32(ekPubKey.(*rsa.PublicKey).E)
	pub.RSAParameters.ModulusRaw = ekPubKey.(*rsa.PublicKey).N.Bytes()
	return &pub
}

func Tpm2MakeCredential(ekPubKey crypto.PublicKey, credential, name []byte) ([]byte, []byte, error) {
	simulatorMutex.Lock()
	defer simulatorMutex.Unlock()

	simulator, err := simulator.Get()
	if err != nil {
		return nil, nil, errors.New("failed get the simulator")
	}
	defer simulator.Close()

	ekPub := pubKeyToTPMPublic(ekPubKey)
	protectHandle, _, err := tpm2.LoadExternal(simulator, *ekPub, tpm2.Private{}, tpm2.HandleNull)
	if err != nil {
		return nil, nil, errors.New("failed load ekPub")
	}

	//generate the credential
	encKeyBlob, encSecret, err := tpm2.MakeCredential(simulator, protectHandle, credential, name)
	if err != nil {
		return nil, nil, errors.New("failed the MakeCredential")
	}

	return encKeyBlob, encSecret, nil
}

func TestMakeCredential(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, RsaKeySize)
	if err != nil {
		t.Fatalf(strPRIVERR, err)
	}
	var testCases = []struct {
		pubkey     crypto.PublicKey
		credential []byte
		name       []byte
	}{
		{&priv.PublicKey, []byte("abc"), []byte("defad")},
	}
	for _, tc := range testCases {
		a1, b1, err1 := MakeCredential(tc.pubkey, tc.credential, tc.name)
		a2, b2, err2 := Tpm2MakeCredential(tc.pubkey, tc.credential, tc.name)
		if err1 != nil || err2 != nil || !bytes.Equal(a1, a2) || !bytes.Equal(b1, b2) {
			//t.Errorf("blob & secret can't match:\n (%v, %v)\n (%v, %v)\n", a1, b1, a2, b2)
			t.Logf("blob & secret can't match:\n (%v, %v)\n (%v, %v)\n", a1, b1, a2, b2)
		}
	}
}