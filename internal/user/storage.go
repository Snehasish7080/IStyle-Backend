package user

import (
	"context"
	"time"

	"errors"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/zone/IStyle/internal/models"
	"github.com/zone/IStyle/pkg/jwtclaim"
	"github.com/zone/IStyle/pkg/otp"
)

type UserStorage struct {
	db     neo4j.DriverWithContext
	dbName string
}

func NewUserStorage(db neo4j.DriverWithContext, dbName string) *UserStorage {
	return &UserStorage{
		db:     db,
		dbName: dbName,
	}
}

func (u *UserStorage) signUp(firstName string, lastName string, userName string, email string, ctx context.Context) (string, error) {

	now := time.Now()
	isEmailExist := u.emailExists(email, ctx)

	if isEmailExist {
		return "", errors.New("email already exists")
	}
	isUserNameExist := u.userNameExists(userName, ctx)

	if isUserNameExist {
		return "", errors.New("username already exists")
	}

	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	generatedOtp := otp.EncodeToString(6)
	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx,
				"CREATE (:User {firstName: $firstName, lastName: $lastName, userName: $userName, email: $email, otp:$otp, isVerified:$isVerified, isComplete:$isComplete, created_at:datetime($createdAt), updated_at:datetime($updatedAt)})",
				map[string]any{"firstName": firstName, "lastName": lastName, "userName": userName, "email": email, "otp": generatedOtp, "isVerified": false, "isComplete": false, "createdAt": now.Format(time.RFC3339), "updatedAt": now.Format(time.RFC3339)})
		})

	if err != nil {
		return "", err
	}

	verifyToken, err := jwtclaim.CreateJwtToken(userName, false)

	if err != nil {
		return "", err
	}

	return verifyToken, nil

}

func (u *UserStorage) verify(otp string, userName string, ctx context.Context) (string, error) {
	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx,
				"MATCH (u:User {userName:$userName}) RETURN u.otp AS otp",
				map[string]interface{}{
					"userName": userName,
				},
			)
			if err != nil {
				return "", err
			}

			record, err := result.Single(ctx)
			if err != nil {
				return "", err
			}

			otp, _ := record.Get("otp")
			return otp.(string), nil

		})

	if err != nil {
		return "", err
	}

	if result != otp {
		return "", errors.New("invalid otp")
	}

	_, err = session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx,
				"MATCH (u:User {userName:$userName}) SET u.isVerified = true",
				map[string]interface{}{
					"userName": userName,
				},
			)
		},
	)
	if err != nil {
		return "", err
	}
	verifyToken, err := jwtclaim.CreateJwtToken(userName, true)
	if err != nil {
		return "", err
	}
	return verifyToken, nil
}

func (u *UserStorage) login(mobile string, ctx context.Context) (string, error) {
	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	isMobileExist := u.mobileExists(mobile, ctx)

	if !isMobileExist {
		return "", errors.New("mobile not registered")
	}

	result, _ := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			result, err := tx.Run(ctx,
				"MATCH (u:User {mobile:$mobile}) RETURN u.userName AS userName",
				map[string]interface{}{
					"mobile": mobile,
				},
			)
			if err != nil {
				return nil, err
			}
			record, err := result.Single(ctx)
			if err != nil {
				return nil, err
			}
			userName, _ := record.Get("userName")
			return userName.(string), nil
		})

	userName, convErr := result.(string)

	if !convErr {
		return "", errors.New("not able to covert")
	}

	generatedOtp := otp.EncodeToString(4)
	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx,
				"MATCH (u:User {userName:$userName}) SET u.isVerified=false, u.otp=$otp",
				map[string]interface{}{
					"userName": userName,
					"otp":      generatedOtp,
				},
			)
		},
	)

	if err != nil {
		return "", err
	}

	verifyToken, err := jwtclaim.CreateJwtToken(userName, false)
	if err != nil {
		return "", err
	}
	return verifyToken, nil
}

func (u *UserStorage) getUser(userName string, ctx context.Context) (*models.User, error) {
	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, _ := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			result, err := tx.Run(ctx,
				"MATCH (u:User {userName:$userName}) RETURN u.firstName AS firstName, u.lastName AS lastName, u.userName AS userName",
				map[string]interface{}{
					"userName": userName,
				},
			)
			if err != nil {
				return nil, err
			}

			record, err := result.Single(ctx)
			if err != nil {
				return nil, err
			}
			firstName, _ := record.Get("firstName")
			lastName, _ := record.Get("lastName")
			userName, _ := record.Get("userName")
			return &models.User{
				FirstName: firstName.(string),
				LastName:  lastName.(string),
				UserName:  userName.(string),
			}, nil
		})

	user, err := result.(*models.User)

	if !err {
		return nil, errors.New("not able to convert")
	}

	return user, nil

}

func (u *UserStorage) updateUser(userName string, userField map[string]interface{}, ctx context.Context) (string, error) {
	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx,
				"MATCH (u:User {userName:$userName}) SET u+=$fields",
				map[string]interface{}{
					"userName": userName,
					"fields":   userField,
				},
			)
		},
	)

	if err != nil {
		return "", err
	}

	return "Update Successfully", nil
}

func (u *UserStorage) mobileExists(mobile string, ctx context.Context) bool {
	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, _ := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			result, err := tx.Run(ctx,
				"MATCH (u:User {mobile:$mobile}) RETURN u.mobile AS mobile",
				map[string]interface{}{
					"mobile": mobile,
				},
			)
			if err != nil {
				return nil, err
			}
			record, err := result.Single(ctx)
			if err != nil {
				return nil, err
			}
			mobile, _ := record.Get("mobile")
			return mobile.(string), nil
		})

	return result != nil

}

func (u *UserStorage) emailExists(email string, ctx context.Context) bool {
	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, _ := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			result, err := tx.Run(ctx,
				"MATCH (u:User {email:$email}) RETURN u.email AS email",
				map[string]interface{}{
					"email": email,
				},
			)
			if err != nil {
				return nil, err
			}
			record, err := result.Single(ctx)
			if err != nil {
				return nil, err
			}
			email, _ := record.Get("email")
			return email.(string), nil
		})

	return result != nil

}

func (u *UserStorage) userNameExists(userName string, ctx context.Context) bool {
	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, _ := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			result, err := tx.Run(ctx,
				"MATCH (u:User {userName:$userName}) RETURN u.userName AS userName",
				map[string]interface{}{
					"userName": userName,
				},
			)
			if err != nil {
				return nil, err
			}
			record, err := result.Single(ctx)
			if err != nil {
				return nil, err
			}
			userName, _ := record.Get("userName")
			return userName.(string), nil
		})

	return result != nil

}
