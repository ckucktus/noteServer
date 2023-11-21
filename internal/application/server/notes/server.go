package notes

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"test_task/internal/domain/entity"
	"test_task/internal/infrastructure/integration"
)

type NoteStorage interface {
	SaveNote(NoteText string, authorId int64) (int64, error)
	GetNotes(AuthorId int64) ([]entity.Note, error)
}

type userStorage interface {
	GetUser(login string) (entity.User, error)
}

type TextChecker interface {
	Validate(ctx context.Context, text string) ([]integration.ValidateTextResponse, error)
}

type Server struct {
	noteStorage NoteStorage
	userStorage userStorage
	textChecker TextChecker
}

func NewServer(noteStorage NoteStorage, userStorage userStorage, textChecker TextChecker) Server {
	return Server{
		noteStorage: noteStorage,
		userStorage: userStorage,
		textChecker: textChecker,
	}
}

type CreateNoteRequest struct {
	Text      string `json:"text"`
	UserLogin string `json:"UserLogin"`
}

type ErrorWord struct {
	Word         string   `json:"word"`
	Alternatives []string `json:"alternatives"`
}

type TextValidationErr struct {
	Errors []ErrorWord `json:"errors"`
}

func (s Server) PostV1CreateNote(ctx echo.Context) (err error) {
	var req CreateNoteRequest

	if err = ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Unwrap(err).Error())
	}
	user, err := s.userStorage.GetUser(req.UserLogin)
	if err != nil {
		return err
	}

	validateTextResponse, err := s.textChecker.Validate(ctx.Request().Context(), req.Text)
	if err != nil {
		return err
	}
	if len(validateTextResponse) > 0 {
		resp := TextValidationErr{Errors: make([]ErrorWord, len(validateTextResponse))}

		for i, errorWord := range validateTextResponse {
			resp.Errors[i].Word = errorWord.Word
			resp.Errors[i].Alternatives = errorWord.S
		}
		return ctx.JSON(http.StatusBadRequest, resp)
	}

	_, err = s.noteStorage.SaveNote(req.Text, user.Id)
	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}

type ListNotesRequest struct {
	UserLogin string `query:"UserLogin"`
}
type Note struct {
	noteText string
}
type ListNotesResponse struct {
	Notes []entity.Note
}

func (s Server) GetV1ListNotes(ctx echo.Context) (err error) {
	var req ListNotesRequest

	if err = ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Unwrap(err).Error())
	}
	user, err := s.userStorage.GetUser(req.UserLogin)
	if err != nil {
		return err
	}

	notes, err := s.noteStorage.GetNotes(user.Id)
	listNotes := ListNotesResponse{Notes: notes}
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, listNotes)
}
