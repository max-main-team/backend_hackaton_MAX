package repositories

import (
	"context"
	"time"

	"github.com/max-main-team/backend_hackaton_MAX/internal/models"
	personalities2 "github.com/max-main-team/backend_hackaton_MAX/internal/models/http/personalities"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models/repository/personalities"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models/repository/schedules"
	"github.com/max-main-team/backend_hackaton_MAX/internal/models/repository/subjects"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	GetUserRolesByID(ctx context.Context, id int64) (*models.UserRoles, error)
	CreateNewUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
}

type RefreshTokenRepository interface {
	Save(token *models.RefreshToken) error
	Find(tokenString string) (*models.RefreshToken, error)
	Delete(tokenString string) error
	DeleteByUser(userID int) error
}

type UniRepository interface {
	GetAllUniversities(ctx context.Context) ([]models.UniversitiesData, error)

	CreateSemestersForUniversity(ctx context.Context, uniID int64, periods []models.SemesterPeriod) error

	GetUniInfoByUserID(ctx context.Context, id int64) (*models.UniversitiesData, error)

	CreateNewDepartment(ctx context.Context, departmentName, departmentCode, aliasName string, facultyID, universityID int64) error

	CreateNewCourse(ctx context.Context, startDate, endDate time.Time, universityDepartmentID int64) error

	GetAllCoursesByUniversityID(ctx context.Context, universityID int64) ([]models.Course, error)

	CreateNewGroup(ctx context.Context, groupName string, courseID int64) error

	CreateNewEvent(ctx context.Context, event models.Event) error

	GetAllEventsByUniversityID(ctx context.Context, universityID int64) ([]models.Event, error)
}

type PersonalitiesRepository interface {
	RequestUniversityAccess(ctx context.Context, uniAccess personalities.UniversityAccess) error
	GetAccessRequest(ctx context.Context, userID, limit, offset int64) (personalities.AccessRequests, error)
	AddNewUser(ctx context.Context, request personalities2.AcceptAccessRequest) error
	DeleteRequest(ctx context.Context, requestID int64) error
	GetAllUniversitiesForPerson(ctx context.Context, userID int64) ([]models.UniversitiesData, error)
	GetAllFacultiesForUniversity(ctx context.Context, universityID int64) ([]models.Faculties, error)
	GetAllDepartmentsForFaculty(ctx context.Context, facultyID int64) ([]models.Departments, error)
	GetAllGroupsForDepartment(ctx context.Context, departmentID int64) ([]models.Groups, error)
	GetAllStudentsForGroup(ctx context.Context, groupID int64) ([]models.User, error)
	GetAllTeachersForUniversity(ctx context.Context, universityID int64) ([]models.User, error)
}

type FaculRepository interface {
	GetFaculsByUserID(ctx context.Context, id int64) ([]models.Faculties, error)
	CreateFaculty(ctx context.Context, id int64, facultyName string) error
}

type SubjectsRepository interface {
	Create(ctx context.Context, name string, uniID int64) error
	Get(ctx context.Context, uniID, limit, offset int64) (*subjects.Subjects, error)
	Delete(ctx context.Context, id int64) error
}

type SchedulesRepository interface {
	CreateClass(ctx context.Context, class schedules.Class) (int64, error)
	DeleteClass(ctx context.Context, classID int64) error
	GetClassesByUniversity(ctx context.Context, universityID int64) ([]schedules.Class, error)

	CreateRoom(ctx context.Context, room schedules.Room) (int64, error)
	DeleteRoom(ctx context.Context, roomID int64) error
	GetRoomsByUniversity(ctx context.Context, universityID int64) ([]schedules.Room, error)

	CreateLesson(ctx context.Context, req schedules.CreateLesson) (int64, error)
	DeleteLesson(ctx context.Context, lessonID int64) error
	GetUserSchedule(ctx context.Context, userID int64) ([]schedules.UserScheduleItem, error)
}
