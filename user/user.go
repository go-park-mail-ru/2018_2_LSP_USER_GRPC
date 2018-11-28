package user

import (
	"database/sql"

	"github.com/go-park-mail-ru/2018_2_LSP_USER_GRPC/user_proto"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type User struct {
	ID         int64
	Email      string
	Username   string
	TotalScore int64
	TotalGames int64
	Password   int64

	FirstName *string
	LastName  *string
	Avatar    *string
}

type UserManager struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func NewUserManager(db *sql.DB, logger *zap.SugaredLogger) *UserManager {
	return &UserManager{db, logger}
}

func (sm *UserManager) GetOne(ctx context.Context, in *user_proto.UserID) (*user_proto.User, error) {
	rows, err := sm.db.Query("SELECT id, username, email, firstname, lastname, avatar, totalgames, totalscore, password FROM users WHERE id = $1", in.ID)
	if err != nil {
		sm.logger.Fatalw("Internal error",
			"err", err,
		)
		return nil, grpc.Errorf(codes.Internal, "Internal server error")
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, grpc.Errorf(codes.NotFound, "User not found")
	}
	u := user_proto.User{}
	FirstName := sql.NullString{}
	LastName := sql.NullString{}
	Avatar := sql.NullString{}
	err = rows.Scan(&u.ID, &u.Username, &u.Email, &FirstName, &LastName, &Avatar, &u.TotalGames, &u.TotalScore, &u.Password)
	if err != nil {
		sm.logger.Fatalw("Internal error",
			"err", err,
		)
		return nil, grpc.Errorf(codes.Internal, "Internal server error")
	}
	if FirstName.Valid {
		u.FirstName = FirstName.String
	}
	if LastName.Valid {
		u.LastName = LastName.String
	}
	if Avatar.Valid {
		u.Avatar = Avatar.String
	}
	return &u, nil
}

func (sm *UserManager) GetOneByEmail(ctx context.Context, in *user_proto.UserEmail) (*user_proto.User, error) {
	rows, err := sm.db.Query("SELECT id, username, email, firstname, lastname, avatar, totalgames, totalscore, password FROM users WHERE email = $1", in.Email)
	if err != nil {
		sm.logger.Fatalw("Internal error",
			"err", err,
		)
		return nil, grpc.Errorf(codes.Internal, "Internal server error")
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, grpc.Errorf(codes.NotFound, "User not found")
	}
	u := user_proto.User{}
	FirstName := sql.NullString{}
	LastName := sql.NullString{}
	Avatar := sql.NullString{}
	err = rows.Scan(&u.ID, &u.Username, &u.Email, &FirstName, &LastName, &Avatar, &u.TotalGames, &u.TotalScore, &u.Password)
	if err != nil {
		sm.logger.Fatalw("Internal error",
			"err", err,
		)
		return nil, grpc.Errorf(codes.Internal, "Internal server error")
	}
	if FirstName.Valid {
		u.FirstName = FirstName.String
	}
	if LastName.Valid {
		u.LastName = LastName.String
	}
	if Avatar.Valid {
		u.Avatar = Avatar.String
	}
	return &u, nil
}

func (sm *UserManager) GetMany(in *user_proto.ManyUsersOptions, stream user_proto.UserChecker_GetManyServer) error {
	rows, err := sm.db.Query("SELECT id, username, email, firstname, lastname, avatar, totalgames, totalscore FROM users ORDER BY "+in.OrderBy+" DESC LIMIT 10 OFFSET $1", in.Page*10)
	if err != nil {
		sm.logger.Fatalw("Internal error",
			"err", err,
		)
		return grpc.Errorf(codes.Internal, err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		u := user_proto.User{}
		FirstName := sql.NullString{}
		LastName := sql.NullString{}
		Avatar := sql.NullString{}
		err = rows.Scan(&u.ID, &u.Username, &u.Email, &FirstName, &LastName, &Avatar, &u.TotalGames, &u.TotalScore)
		if err != nil {
			sm.logger.Fatalw("Internal error",
				"err", err,
			)
			return grpc.Errorf(codes.Internal, err.Error())
		}
		if FirstName.Valid {
			u.FirstName = FirstName.String
		}
		if LastName.Valid {
			u.LastName = LastName.String
		}
		if Avatar.Valid {
			u.Avatar = Avatar.String
		}
		err = stream.Send(&u)
		if err != nil {
			sm.logger.Fatalw("Internal error",
				"err", err,
			)
			return grpc.Errorf(codes.Internal, err.Error())
		}
	}

	return nil
}

func (sm *UserManager) Create(ctx context.Context, in *user_proto.User) (*user_proto.UserID, error) {
	err := sm.validateRegisterUnique(in)
	if err != nil {
		return nil, err
	}

	hashed, err := hashPassword(in.Password)
	if err != nil {
		sm.logger.Fatalw("Internal error",
			"err", err,
		)
		return nil, grpc.Errorf(codes.Internal, "Internal server error")
	}

	rows, err := sm.db.Query("INSERT INTO users (firstname, lastname, email, password, username) VALUES ($1, $2, $3, $4, $5) RETURNING id;", in.FirstName, in.LastName, in.Email, hashed, in.Username)
	if err != nil {
		sm.logger.Fatalw("Internal error",
			"err", err,
		)
		return nil, grpc.Errorf(codes.Internal, "Internal server error")
	}
	defer rows.Close()

	u := user_proto.UserID{}

	rows.Next()
	err = rows.Scan(&u.ID)
	if err != nil {
		sm.logger.Fatalw("Internal error",
			"err", err,
		)
		return nil, grpc.Errorf(codes.Internal, "Internal server error")
	}
	return &u, nil
}

func (sm *UserManager) Update(ctx context.Context, in *user_proto.User) (*user_proto.UserID, error) {
	request := "UPDATE users SET "
	hasToUpdate := false
	if len(in.Username) > 0 {
		request += "username='" + in.Username + "',"
		hasToUpdate = true
	}
	if len(in.FirstName) > 0 {
		request += "firstname='" + in.FirstName + "',"
		hasToUpdate = true
	}
	if len(in.LastName) > 0 {
		request += "lastname='" + in.LastName + "',"
		hasToUpdate = true
	}
	if len(in.Email) > 0 {
		request += "email='" + in.Email + "',"
		hasToUpdate = true
	}
	if len(in.Avatar) > 0 {
		request += "avatar='" + in.Avatar + "',"
		hasToUpdate = true
	}
	if len(in.Password) > 0 {
		hashed, err := hashPassword(in.Password)
		if err != nil {
			sm.logger.Fatalw("Internal error",
				"err", err,
			)
			return nil, grpc.Errorf(codes.Internal, "Internal server error")
		}
		request += "password='" + hashed + "',"
		hasToUpdate = true
	}
	if !hasToUpdate {
		return nil, grpc.Errorf(codes.InvalidArgument, "You must specify fields to update")
	}
	request = request[:len(request)-1]
	request += " WHERE id = $1 RETURNING id"

	rows, err := sm.db.Query(request, in.ID)
	if err != nil {
		sm.logger.Fatalw("Internal error",
			"err", err,
		)
		return nil, grpc.Errorf(codes.Internal, "Internal server error")
	}
	defer rows.Close()

	u := user_proto.UserID{}
	if !rows.Next() {
		return nil, grpc.Errorf(codes.NotFound, "User not found")
	}
	err = rows.Scan(&u.ID)
	if err != nil {
		sm.logger.Fatalw("Internal error",
			"err", err,
		)
		return nil, grpc.Errorf(codes.Internal, "Internal server error")
	}

	return &u, nil
}

func (sm *UserManager) validateRegisterUnique(u *user_proto.User) error {
	rows, err := sm.db.Query("SELECT EXISTS (SELECT * FROM users WHERE email = $1 LIMIT 1) AS email, EXISTS (SELECT * FROM users WHERE username = $2 LIMIT 1) AS username", u.Email, u.Username)
	if err != nil {
		sm.logger.Fatalw("Internal error",
			"err", err,
		)
		return grpc.Errorf(codes.Internal, "Internal server error")
	}
	defer rows.Close()

	if !rows.Next() {
		sm.logger.Fatalw("Internal error",
			"err", err,
		)
		return grpc.Errorf(codes.Internal, "Internal server error")
	}

	emailTaken, usernameTaken := false, false
	err = rows.Scan(&emailTaken, &usernameTaken)
	if err != nil {
		sm.logger.Fatalw("Internal error",
			"err", err,
		)
		return grpc.Errorf(codes.Internal, "Internal server error")
	}

	if emailTaken {
		return grpc.Errorf(codes.AlreadyExists, "Email is already taken")
	}
	if usernameTaken {
		return grpc.Errorf(codes.AlreadyExists, "Username is already taken")
	}

	return nil
}
