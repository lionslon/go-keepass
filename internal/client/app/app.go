package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/lionslon/go-keepass/internal/client/config"
	"github.com/lionslon/go-keepass/internal/crypt"
	"github.com/lionslon/go-keepass/internal/models"
	"net/http"
	"strings"
)

const (
	registerUrl = "api/user/register"
	loginUrl    = "api/user/login"
	addDataUrl  = "api/data"
)

// sender для взаимодействия клиента с сервером
type sender struct {
	cfg       *config.Config   // конфиг приложения
	client    *resty.Client    // клиент http
	encryptor *crypt.Encryptor // объект для шифрования аутентификационных данных на открытом ключе сервера
	token     string           // актуальный jwt токен
	password  string           // пароль пользователя (используем для шифрования / расшифровывания данных для /от сервера)
}

func NewSender(cfg *config.Config) sender {

	if !strings.HasPrefix(cfg.ServerEndpoint, "http") && !strings.HasPrefix(cfg.ServerEndpoint, "https") {
		cfg.ServerEndpoint = "http://" + cfg.ServerEndpoint
	}

	return sender{
		cfg:    cfg,
		client: resty.New(),
	}
}

func (m *sender) Init() error {

	encryptor, err := crypt.NewEncryptor(m.cfg.CryptoKey)
	if err != nil {
		return fmt.Errorf("cannot create credentions encryptor: %w", err)
	}

	m.encryptor = encryptor

	return nil
}

func (m *sender) Register(login, password string) error {

	encryptAuthData, err := m.createEncryptUserAuthData(login, password)
	if err != nil {
		return fmt.Errorf("cannot create encrypt user auth data: %w", err)
	}

	req := m.client.R().SetBody(encryptAuthData)

	url := strings.Join([]string{m.cfg.ServerEndpoint, registerUrl}, "/")

	resp, err := req.Post(url)
	if err != nil {
		return fmt.Errorf("cannot send register request: %w", err)
	}

	//Нужно разобрать заголовки и забрать токен
	if code := resp.StatusCode(); code != http.StatusOK {
		return fmt.Errorf("request processing failed, code: %d", code)
	}

	if err := m.parseAuthorization(resp); err != nil {
		return fmt.Errorf("cannot get jwt token: %s", err)
	}

	m.password = password

	return nil
}

func (m *sender) Login(login, password string) error {

	encryptAuthData, err := m.createEncryptUserAuthData(login, password)
	if err != nil {
		return fmt.Errorf("cannot create encrypt user auth data: %w", err)
	}

	req := m.client.R().SetBody(encryptAuthData)

	url := strings.Join([]string{m.cfg.ServerEndpoint, loginUrl}, "/")

	resp, err := req.Post(url)
	if err != nil {
		return fmt.Errorf("cannot send login request: %w", err)
	}

	//Нужно разобрать заголовки и забрать токен
	if code := resp.StatusCode(); code != http.StatusOK {
		return fmt.Errorf("request processing failed, code: %d", code)
	}

	if err := m.parseAuthorization(resp); err != nil {
		return fmt.Errorf("cannot get jwt token: %s", err)
	}

	m.password = password

	return nil
}

func (m *sender) AddNewData(identifier string, data []byte) error {

	if m.token == `` || m.password == `` {
		return fmt.Errorf("bad auth data, try login")
	}

	encryptData, err := crypt.SymmetricEncrypt(m.password, data)
	if err != nil {
		return fmt.Errorf("cannot encrypt user data: %w", err)
	}

	req := m.client.R().
		SetBody(encryptData).
		SetHeader("Authorization", m.token)

	url := strings.Join([]string{m.cfg.ServerEndpoint, addDataUrl, identifier}, "/")

	resp, err := req.Post(url)
	if err != nil {
		return fmt.Errorf("cannot send add data request: %w", err)
	}

	if code := resp.StatusCode(); code != http.StatusAccepted {
		return fmt.Errorf("request processing failed, code: %d", code)
	}

	return nil
}

func (m *sender) GetUserData(identifier string) ([]byte, error) {
	if m.token == `` || m.password == `` {
		return nil, fmt.Errorf("bad auth data, try login")
	}

	req := m.client.R().
		SetHeader("Authorization", m.token)

	url := strings.Join([]string{m.cfg.ServerEndpoint, addDataUrl, identifier}, "/")

	resp, err := req.Get(url)
	if err != nil {
		return nil, fmt.Errorf("cannot sen get user data request: %w", err)
	}

	if code := resp.StatusCode(); code != http.StatusOK {
		return nil, fmt.Errorf("request processing failed, code: %d", code)
	}

	data, err := crypt.SymmetricDecrypt(m.password, resp.Body())
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt user data: %w", err)
	}

	return data, nil
}

// Шифрует аутентификационные данные пользователя
func (m *sender) createEncryptUserAuthData(login, password string) ([]byte, error) {
	authdto := new(bytes.Buffer)
	if err := json.NewEncoder(authdto).Encode(&models.AuthDTO{
		Login:    login,
		Password: password,
	}); err != nil {
		return nil, fmt.Errorf("error encoding auth dto %w", err)
	}

	//Шифруем аутентификационные данные
	encryptbuf, err := m.encryptor.Encrypt(authdto.Bytes())
	if err != nil {
		return nil, fmt.Errorf("cannot encrypt user auth data: %w", err)
	}

	return encryptbuf, nil
}

func (m *sender) parseAuthorization(resp *resty.Response) error {

	resp.Header().Get("Authorization")

	//Получение header c токеном
	m.token = resp.Header().Get("Authorization")
	if m.token == `` {
		return fmt.Errorf("authorization header is missing")
	}

	return nil
}
