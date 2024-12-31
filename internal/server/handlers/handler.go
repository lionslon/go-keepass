package handlers

import (
	"fmt"
	"github.com/lionslon/go-keepass/internal/auth"
	"github.com/lionslon/go-keepass/internal/logger"
	"github.com/lionslon/go-keepass/internal/models"
	"github.com/lionslon/go-keepass/internal/storage"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type KeeperHandler struct {
	storage *storage.KeeperStorage
}

func NewKeeperHandler(storage *storage.KeeperStorage) KeeperHandler {
	return KeeperHandler{
		storage: storage,
	}
}

func (m *KeeperHandler) Register(r *chi.Mux) {

	r.Route("/api/user", func(r chi.Router) {
		//Регистрация нового пользователя
		r.Post("/register", m.userRegister)
		//Аутентификация существующего пользователя
		r.Post("/login", m.login)
	})
}

func (m *KeeperHandler) errorRespond(w http.ResponseWriter, code int, err error) {
	logger.Error(err.Error())
	w.WriteHeader(code)
}

func (m *KeeperHandler) userRegister(w http.ResponseWriter, r *http.Request) {

	//Разобрали запрос
	authDTO, err := models.NewDTO[models.AuthDTO](r.Body)
	if err != nil {
		m.errorRespond(w, http.StatusBadRequest, fmt.Errorf("cannot decode auth dto: %s", err))
		return
	}
	//Проверили наличие полей
	if err := authDTO.Validate(); err != nil {
		m.errorRespond(w, http.StatusBadRequest, fmt.Errorf("cannot validate auth dto: %s", err))
		return
	}
	//Проверяем, что пользака с таким логином нет
	if m.storage.IsUserExist(r.Context(), authDTO.Login) {
		m.errorRespond(w, http.StatusConflict, fmt.Errorf("user with login %s already exist", authDTO.Login))
		return
	}
	//Создаем пользователя, получаем идентификатор для токена
	user_id, err := m.storage.CreateUser(r.Context(), authDTO)
	if err != nil {
		m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot create new user: %s", err))
		return
	}

	//Выпускаем токен, посылаем в заголовке ответа
	jwt, err := auth.CreateToken(user_id)
	if err != nil {
		m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot create jwt: %s", err))
		return
	}

	w.Header().Set("Authorization", jwt)
}

func (m *KeeperHandler) login(w http.ResponseWriter, r *http.Request) {
	//Разобрали запрос
	authDTO, err := models.NewDTO[models.AuthDTO](r.Body)
	if err != nil {
		m.errorRespond(w, http.StatusBadRequest, fmt.Errorf("cannot decode auth dto: %s", err))
		return
	}

	//Провереяем корректность данных пользователя
	user_id, err := m.storage.Login(r.Context(), authDTO)
	if err != nil {
		m.errorRespond(w, http.StatusUnauthorized, fmt.Errorf("authentication failed: %s", err))
		return
	}

	//Выпускаем токен, посылаем в заголовке ответа
	jwt, err := auth.CreateToken(user_id)
	if err != nil {
		m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot create jwt: %s", err))
		return
	}

	w.Header().Set("Authorization", jwt)
}
