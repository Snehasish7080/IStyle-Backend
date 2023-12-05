package user

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/zone/IStyle/internal/models"
	"github.com/zone/IStyle/pkg/hash"
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

func (u *UserStorage) signUp(firstName string, lastName string, userName string, email string, password string, ctx context.Context) (string, error) {
	now := time.Now()
	isEmailExist := u.emailExists(email, ctx)

	if isEmailExist {
		return "", errors.New("email already exists")
	}
	isUserNameExist := u.userNameExists(userName, ctx)

	if isUserNameExist {
		return "", errors.New("username already exists")
	}

	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		return "", err
	}

	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	generatedOtp := otp.EncodeToString(6)
	_, err = session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx,
				"CREATE (:User {firstName: $firstName, lastName: $lastName, userName: $userName, email: $email, password: $password, emailOtp:$emailOtp, isEmailVerified:$isEmailVerified, isMobileVerified:$isMobileVerified, isComplete:$isComplete, created_at:datetime($createdAt), updated_at:datetime($updatedAt)})",
				map[string]any{"firstName": firstName, "lastName": lastName, "userName": userName, "email": email, "password": hashedPassword, "emailOtp": generatedOtp, "isEmailVerified": false, "isMobileVerified": false, "isComplete": false, "createdAt": now.Format(time.RFC3339), "updatedAt": now.Format(time.RFC3339)})
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

func (u *UserStorage) verifyEmail(otp string, userName string, ctx context.Context) (string, error) {
	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx,
				"MATCH (u:User {userName:$userName}) RETURN u.emailOtp AS otp",
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
				"MATCH (u:User {userName:$userName}) SET u.isEmailVerified = true",
				map[string]interface{}{
					"userName": userName,
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

func (u *UserStorage) verifyMobile(otp string, userName string, ctx context.Context) (string, error) {
	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx,
				"MATCH (u:User {userName:$userName}) RETURN u.mobileOtp AS otp",
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
				"MATCH (u:User {userName:$userName}) SET u.isMobileVerified = true",
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

func (u *UserStorage) login(email string, password string, ctx context.Context) (string, error) {
	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	isEmailExist := u.emailExists(email, ctx)

	if !isEmailExist {
		return "", errors.New("email not registered")
	}

	result, _ := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			result, err := tx.Run(ctx,
				"MATCH (u:User {email:$email}) RETURN u.userName AS userName, u.password AS password, u.isMobileVerified AS isMobileVerified",
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
			userName, _ := record.Get("userName")
			password, _ := record.Get("password")
			isMobileVerified, _ := record.Get("isMobileVerified")
			return &models.User{
				UserName:         userName.(string),
				Password:         password.(string),
				IsMobileVerified: isMobileVerified.(bool),
			}, nil
		})

	user, convErr := result.(*models.User)

	if !convErr {
		return "", errors.New("not able to covert")
	}

	if !hash.CheckPasswordHash(password, user.Password) {
		return "", errors.New("incorrect email or password")
	}

	verifyToken, err := jwtclaim.CreateJwtToken(user.UserName, user.IsMobileVerified)
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
				"MATCH (u:User {userName:$userName}) RETURN u.firstName AS firstName, u.lastName AS lastName, u.userName AS userName, u.bio AS bio, u.profilePic AS profilePic, u.isMobileVerified AS isMobileVerified, u.isComplete AS isComplete",
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
			bio, _ := record.Get("bio")
			profilePic, _ := record.Get("profilePic")
			isMobileVerified, _ := record.Get("isMobileVerified")
			isComplete, _ := record.Get("isComplete")
			if bio == nil {
				bio = ""
			}
			if profilePic == nil {
				profilePic = ""
			}
			return &models.User{
				FirstName:        firstName.(string),
				LastName:         lastName.(string),
				UserName:         userName.(string),
				Bio:              bio.(string),
				ProfilePic:       profilePic.(string),
				IsMobileVerified: isMobileVerified.(bool),
				IsComplete:       isComplete.(bool),
			}, nil
		})

	user, err := result.(*models.User)

	if !err {
		return nil, errors.New("not able to convert")
	}

	return user, nil
}

func (u *UserStorage) updateMobile(userName string, mobile string, otp string, ctx context.Context) (string, error) {
	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	isMobileExist := u.mobileExists(mobile, ctx)

	if isMobileExist {
		return "", errors.New("mobile already exists")
	}

	now := time.Now()
	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx,
				"MATCH (u:User {userName:$userName}) SET u.updated_at=datetime($updatedAt), u.mobile=$mobile, u.mobileOtp=$mobileOtp",
				map[string]interface{}{
					"userName":  userName,
					"updatedAt": now.Format(time.RFC3339),
					"mobile":    mobile,
					"mobileOtp": otp,
				},
			)
		},
	)
	if err != nil {
		return "", err
	}

	return "Update Successfully", nil
}

func (u *UserStorage) updateUser(userName string, userField map[string]interface{}, ctx context.Context) (string, error) {
	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	now := time.Now()
	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx,
				"MATCH (u:User {userName:$userName}) SET u.updated_at=datetime($updatedAt), (CASE WHEN u.bio = null THEN u END).bio = $fields.bio, (CASE WHEN u.profilePic = null THEN u END).profilePic = $fields.profilePic, u+=$fields",
				map[string]interface{}{
					"userName":  userName,
					"updatedAt": now.Format(time.RFC3339),
					"fields":    userField,
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

func (u *UserStorage) getUserByUserName(userName string, ctx context.Context) (*models.User, error) {
	isUserNameExist := u.userNameExists(userName, ctx)

	if !isUserNameExist {
		return nil, errors.New("user does not exists")
	}
	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, _ := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			result, err := tx.Run(ctx,
				"MATCH (u:User {userName:$userName}) RETURN u.firstName AS firstName, u.lastName AS lastName, u.userName AS userName, u.bio AS bio, u.profilePic AS profilePic",
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
			bio, _ := record.Get("bio")
			profilePic, _ := record.Get("profilePic")
			if bio == nil {
				bio = ""
			}
			if profilePic == nil {
				profilePic = ""
			}
			return &models.User{
				FirstName:  firstName.(string),
				LastName:   lastName.(string),
				UserName:   userName.(string),
				Bio:        bio.(string),
				ProfilePic: profilePic.(string),
			}, nil
		})

	user, err := result.(*models.User)

	if !err {
		return nil, errors.New("not able to convert")
	}

	return user, nil
}

func (u *UserStorage) markFavTags(userName string, tags []string, ctx context.Context) (string, error) {
	session := u.db.NewSession(ctx, neo4j.SessionConfig{DatabaseName: u.dbName, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx,
				`MATCH (u:User {userName:$userName}) 
         UNWIND $tags AS tagId
         MATCH (t:Tag {uuid:tagId})
         MERGE (u)-[:MARK_FAV]->(t)
         SET u.isComplete=true
        `,
				map[string]interface{}{
					"userName": userName,
					"tags":     tags,
				},
			)
		},
	)
	if err != nil {
		return "something went wrong", err
	}

	return "marked successfully", nil
}
