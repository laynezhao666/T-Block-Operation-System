package robot

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"etrpc-go/log"
)

const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	ValidateSignatureError int = -40001
	ParseXmlError          int = -40002
	ComputeSignatureError  int = -40003
	IllegalAesKey          int = -40004
	ValidateCorpidError    int = -40005
	EncryptAESError        int = -40006
	DecryptAESError        int = -40007
	IllegalBuffer          int = -40008
	EncodeBase64Error      int = -40009
	DecodeBase64Error      int = -40010
	GenXmlError            int = -40010
	ParseJsonError         int = -40012
	GenJsonError           int = -40013
	IllegalProtocolType    int = -40014
)

type ProtocolType int

const (
	XmlType ProtocolType = 1
)

// CryptError CryptError
type CryptError struct {
	ErrCode int
	ErrMsg  string
}

// NewCryptError NewCryptError
//
//	@param errCode
//	@param errMsg
//	@return *CryptError
func NewCryptError(errCode int, errMsg string) *CryptError {
	return &CryptError{ErrCode: errCode, ErrMsg: errMsg}
}

// WXBizMsg4Recv WXBizMsg4Recv
type WXBizMsg4Recv struct {
	ToUsername string `xml:"ToUserName"`
	Encrypt    string `xml:"Encrypt"`
	AgentId    string `xml:"AgentID"`
}

// CDATA CDATA
type CDATA struct {
	Value string `xml:",cdata"`
}

// WXBizMsg4Send WXBizMsg4Send
type WXBizMsg4Send struct {
	XMLName   xml.Name `xml:"xml"`
	Encrypt   CDATA    `xml:"Encrypt"`
	Signature CDATA    `xml:"MsgSignature"`
	Timestamp string   `xml:"TimeStamp"`
	Nonce     CDATA    `xml:"Nonce"`
}

// NewWXBizMsg4Send NewWXBizMsg4Send
//
//	@param encrypt
//	@param signature
//	@param timestamp
//	@param nonce
//	@return *WXBizMsg4Send
func NewWXBizMsg4Send(encrypt, signature, timestamp, nonce string) *WXBizMsg4Send {
	return &WXBizMsg4Send{Encrypt: CDATA{Value: encrypt}, Signature: CDATA{Value: signature},
		Timestamp: timestamp, Nonce: CDATA{Value: nonce}}
}

// ProtocolProcessor ProtocolProcessor
type ProtocolProcessor interface {
	parse(srcData []byte) (*WXBizMsg4Recv, *CryptError)
	serialize(msgSend *WXBizMsg4Send) ([]byte, *CryptError)
}

// WXBizMsgCrypt WXBizMsgCrypt
type WXBizMsgCrypt struct {
	token             string
	encodingAesKey    string
	receiverId        string
	protocolProcessor ProtocolProcessor
}

// XmlProcessor XmlProcessor
type XmlProcessor struct {
}

func (o *XmlProcessor) parse(srcData []byte) (*WXBizMsg4Recv, *CryptError) {
	var msg4Recv WXBizMsg4Recv
	err := xml.Unmarshal(srcData, &msg4Recv)
	if nil != err {
		return nil, NewCryptError(ParseXmlError, "xml to msg fail")
	}
	return &msg4Recv, nil
}

func (o *XmlProcessor) serialize(msg4Send *WXBizMsg4Send) ([]byte, *CryptError) {
	xmlMsg, err := xml.Marshal(msg4Send)
	if nil != err {
		return nil, NewCryptError(GenXmlError, err.Error())
	}
	return xmlMsg, nil
}

// NewWXBizMsgCrypt NewWXBizMsgCrypt
//
//	@param token
//	@param encodingAesKey
//	@param receiverId
//	@param protocolType
//	@return *WXBizMsgCrypt
func NewWXBizMsgCrypt(token, encodingAesKey, receiverId string, protocolType ProtocolType) *WXBizMsgCrypt {
	var protocolProcessor ProtocolProcessor
	if protocolType != XmlType {
		panic("unsupport protocal")
	} else {
		protocolProcessor = new(XmlProcessor)
	}
	return &WXBizMsgCrypt{
		token:             token,
		encodingAesKey:    encodingAesKey + "=",
		receiverId:        receiverId,
		protocolProcessor: protocolProcessor,
	}
}

func (o *WXBizMsgCrypt) randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func (o *WXBizMsgCrypt) pKCS7Padding(plaintext string, blockSize int) []byte {
	padding := blockSize - (len(plaintext) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	var buffer bytes.Buffer
	buffer.WriteString(plaintext)
	buffer.Write(padText)
	return buffer.Bytes()
}

func (o *WXBizMsgCrypt) pKCS7UnPadding(plaintext []byte, blockSize int) ([]byte, *CryptError) {
	plaintextLen := len(plaintext)
	if nil == plaintext || plaintextLen == 0 {
		return nil, NewCryptError(DecryptAESError, "pKCS7UnPadding error nil or zero")
	}
	if plaintextLen%blockSize != 0 {
		return nil, NewCryptError(DecryptAESError, "pKCS7UnPadding text not a multiple of the block size")
	}
	paddingLen := int(plaintext[plaintextLen-1])
	return plaintext[:plaintextLen-paddingLen], nil
}

func (o *WXBizMsgCrypt) cbcEncrypt(plaintext string) ([]byte, *CryptError) {
	aesKey, err := base64.StdEncoding.DecodeString(o.encodingAesKey)
	if nil != err {
		return nil, NewCryptError(DecodeBase64Error, err.Error())
	}
	const blockSize = 32
	padMsg := o.pKCS7Padding(plaintext, blockSize)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, NewCryptError(EncryptAESError, err.Error())
	}

	ciphertext := make([]byte, len(padMsg))
	iv := aesKey[:aes.BlockSize]

	mode := cipher.NewCBCEncrypter(block, iv)

	mode.CryptBlocks(ciphertext, padMsg)
	base64Msg := make([]byte, base64.StdEncoding.EncodedLen(len(ciphertext)))
	base64.StdEncoding.Encode(base64Msg, ciphertext)

	return base64Msg, nil
}

func (o *WXBizMsgCrypt) cbcDecrypt(base64EncryptMsg string) ([]byte, *CryptError) {
	aesKey, err := base64.StdEncoding.DecodeString(o.encodingAesKey)
	if nil != err {
		return nil, NewCryptError(DecodeBase64Error, err.Error())
	}

	encryptMsg, err := base64.StdEncoding.DecodeString(base64EncryptMsg)
	if nil != err {
		return nil, NewCryptError(DecodeBase64Error, err.Error())
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, NewCryptError(DecryptAESError, err.Error())
	}

	if len(encryptMsg) < aes.BlockSize {
		return nil, NewCryptError(DecryptAESError, "encrypt_msg size is not valid")
	}

	iv := aesKey[:aes.BlockSize]

	if len(encryptMsg)%aes.BlockSize != 0 {
		return nil, NewCryptError(DecryptAESError, "encrypt_msg not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks(encryptMsg, encryptMsg)

	return encryptMsg, nil
}

func (o *WXBizMsgCrypt) calSignature(timestamp, nonce, data string) string {
	sortArr := []string{o.token, timestamp, nonce, data}
	sort.Strings(sortArr)
	var buffer bytes.Buffer
	for _, value := range sortArr {
		buffer.WriteString(value)
	}

	sha := sha1.New()
	sha.Write(buffer.Bytes())
	signature := fmt.Sprintf("%x", sha.Sum(nil))
	return signature
}

// ParsePlainText ParsePlainText
//
//	@receiver o
//	@param plaintext
//	@return []byte
//	@return uint32
//	@return []byte
//	@return []byte
//	@return *CryptError
func (o *WXBizMsgCrypt) ParsePlainText(plaintext []byte) ([]byte, uint32, []byte, []byte, *CryptError) {
	const blockSize = 32
	plaintext, err := o.pKCS7UnPadding(plaintext, blockSize)
	if nil != err {
		return nil, 0, nil, nil, err
	}

	textLen := uint32(len(plaintext))
	if textLen < 20 {
		return nil, 0, nil, nil, NewCryptError(IllegalBuffer, "plain is to small 1")
	}
	random := plaintext[:16]
	msgLen := binary.BigEndian.Uint32(plaintext[16:20])
	if textLen < (20 + msgLen) {
		return nil, 0, nil, nil, NewCryptError(IllegalBuffer, "plain is to small 2")
	}

	msg := plaintext[20 : 20+msgLen]
	receiverId := plaintext[20+msgLen:]

	return random, msgLen, msg, receiverId, nil
}

// VerifyURL VerifyURL
//
//	@receiver o
//	@param msgSignature
//	@param timestamp
//	@param nonce
//	@param echoStr
//	@return []byte
//	@return *CryptError
func (o *WXBizMsgCrypt) VerifyURL(msgSignature, timestamp, nonce, echoStr string) ([]byte, *CryptError) {
	log.Infof("[verifyURL] msgSignature: %s, timestamp: %s, nonce: %s, echoStr: %s",
		msgSignature, timestamp, nonce, echoStr)
	signature := o.calSignature(timestamp, nonce, echoStr)

	if strings.Compare(signature, msgSignature) != 0 {
		log.Errorf("signature not equal, signature: %s, msgSignature: %s", signature, msgSignature)
		return nil, NewCryptError(ValidateSignatureError, "signature not equal")
	}

	plaintext, err := o.cbcDecrypt(echoStr)
	log.Infof("[verifyURL] plaintext: %s", plaintext)
	if nil != err {
		log.Error("cbcDecrypt Error")
		return nil, err
	}

	_, _, msg, receiverId, err := o.ParsePlainText(plaintext)
	log.Infof("[verifyURL] msg: %s, receiverId: %s", msg, receiverId)
	if nil != err {
		log.Error("ParsePlainText Error")
		return nil, err
	}

	if len(o.receiverId) > 0 && strings.Compare(string(receiverId), o.receiverId) != 0 {
		log.Error("receiver_id is not equil")
		return nil, NewCryptError(ValidateCorpidError, "receiver_id is not equil")
	}

	return msg, nil
}

// EncryptMsg EncryptMsg
//
//	@receiver o
//	@param replyMsg
//	@param timestamp
//	@param nonce
//	@return []byte
//	@return *CryptError
func (o *WXBizMsgCrypt) EncryptMsg(replyMsg, timestamp, nonce string) ([]byte, *CryptError) {
	randStr := o.randString(16)
	var buffer bytes.Buffer
	buffer.WriteString(randStr)

	msgLenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(msgLenBuf, uint32(len(replyMsg)))
	buffer.Write(msgLenBuf)
	buffer.WriteString(replyMsg)
	buffer.WriteString(o.receiverId)

	tmpCiphertext, err := o.cbcEncrypt(buffer.String())
	if nil != err {
		return nil, err
	}
	ciphertext := string(tmpCiphertext)

	signature := o.calSignature(timestamp, nonce, ciphertext)

	msg4Send := NewWXBizMsg4Send(ciphertext, signature, timestamp, nonce)
	return o.protocolProcessor.serialize(msg4Send)
}

// DecryptMsg DecryptMsg
//
//	@receiver o
//	@param msgSignature
//	@param timestamp
//	@param nonce
//	@param postData
//	@return []byte
//	@return *CryptError
func (o *WXBizMsgCrypt) DecryptMsg(msgSignature, timestamp, nonce string, postData []byte) ([]byte, *CryptError) {
	msg4Recv, cryptErr := o.protocolProcessor.parse(postData)
	if nil != cryptErr {
		return nil, cryptErr
	}

	signature := o.calSignature(timestamp, nonce, msg4Recv.Encrypt)

	if strings.Compare(signature, msgSignature) != 0 {
		return nil, NewCryptError(ValidateSignatureError, "signature not equal")
	}

	plaintext, cryptErr := o.cbcDecrypt(msg4Recv.Encrypt)
	if nil != cryptErr {
		return nil, cryptErr
	}

	_, _, msg, receiverId, cryptErr := o.ParsePlainText(plaintext)
	if nil != cryptErr {
		return nil, cryptErr
	}

	if len(o.receiverId) > 0 && strings.Compare(string(receiverId), o.receiverId) != 0 {
		return nil, NewCryptError(ValidateCorpidError, "receiver_id is not equil")
	}

	return msg, nil
}
