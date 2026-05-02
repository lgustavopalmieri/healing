package profile

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/language"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/photo"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/video"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/value_object/academic"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/value_object/services"
)

type Profile struct {
	SpecialistID string
	Id           string
	ProfilePhoto photo.Photo
	Languages    []language.Language
	Photos       []photo.Photo
	Video        video.Video
	Academic     []academic.Academic
	Services     []services.Services
}
