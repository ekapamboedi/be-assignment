package middleware

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"go-be-assignment/config"
	"go-be-assignment/helper"
	"go-be-assignment/model"
)

type AccessPermission struct {
	MainMenu string   `json:"mainMenu"`
	SubMenu  []string `json:"subMenu"`
	Action   []string `json:"action"`
}

var MethodToAction = map[string]string{
	"PUT":    "Edit",
	"DELETE": "Delete",
	"POST":   "Add",
}

func AdminAuthenticate(ctx *gin.Context) {
	var tokenBody helper.TokenPayload
	header := ctx.GetHeader("authorization")

	headerSplit := strings.Split(header, " ")
	tokenText := headerSplit[1]

	token, _ := jwt.Parse(tokenText, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(config.JWT_KEY), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenBody.Role = claims["role"].(string)
	}

	if strings.ToLower(tokenBody.Role) != "admin" {
		ctx.AbortWithStatus(403)
		return
	}

	ctx.Next()
}

func SuperAdminAuthenticate(ctx *gin.Context) {
	var tokenBody helper.TokenPayload
	header := ctx.GetHeader("authorization")

	headerSplit := strings.Split(header, " ")
	tokenText := headerSplit[1]

	token, _ := jwt.Parse(tokenText, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(config.JWT_KEY), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenBody.Role = claims["role"].(string)
	}

	if strings.ToLower(tokenBody.Role) != "super admin" {
		ctx.AbortWithStatus(403)
		return
	}

	ctx.Next()
}

func AuthenticateOnly(ctx *gin.Context) {
	var tokenBody helper.TokenPayload

	defer func() {
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			description := fmt.Sprintf("Panic: %v\n%s\n", err, buf[:n])
			errorMessage := fmt.Sprintf("%v", err)
			fmt.Println(errorMessage, "\n", description)
			go LogErrorWithDescription(ctx.Request.URL.Path, ctx.Request.Method, errorMessage, description, tokenBody)
			ctx.AbortWithStatus(500)
		}
	}()

	header := ctx.GetHeader("authorization")
	if header == "" {
		ctx.AbortWithStatus(401)
		return
	}

	headerSplit := strings.Split(header, " ")
	tokenText := headerSplit[1]

	token, err := jwt.Parse(tokenText, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(config.JWT_KEY), nil
	})

	if err != nil {
		fmt.Println(err)
		ctx.AbortWithStatus(403)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenBody.CompanyId = claims["company_id"].(string)
		tokenBody.Role = claims["role"].(string)
		tokenBody.EmployeeId = claims["employee_id"].(string)
		tokenBody.CompanyName = claims["company_name"].(string)
		tokenBody.Username = claims["username"].(string)
	}

	ctx.Next()
}

func Authenticate(ctx *gin.Context) {
	var tokenBody helper.TokenPayload

	defer func() {
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			description := fmt.Sprintf("Panic: %v\n%s\n", err, buf[:n])
			errorMessage := fmt.Sprintf("%v", err)
			fmt.Println(errorMessage, "\n", description)
			go LogErrorWithDescription(ctx.Request.URL.Path, ctx.Request.Method, errorMessage, description, tokenBody)
			ctx.AbortWithStatus(500)
		}
	}()

	header := ctx.GetHeader("authorization")
	if header == "" {
		ctx.AbortWithStatus(401)
		return
	}

	headerSplit := strings.Split(header, " ")
	tokenText := headerSplit[1]

	token, err := jwt.Parse(tokenText, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(config.JWT_KEY), nil
	})

	if err != nil {
		fmt.Println(err)
		ctx.AbortWithStatus(403)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenBody.CompanyId = claims["company_id"].(string)
		tokenBody.Role = claims["role"].(string)
		tokenBody.EmployeeId = claims["employee_id"].(string)
		tokenBody.CompanyName = claims["company_name"].(string)
		tokenBody.Username = claims["username"].(string)
	}

	var access []AccessPermission
	if strings.ToLower(tokenBody.Role) != "admin" {
		var currentMenuAccess model.MenuAccess
		var currentEmployee model.Employee

		result := model.DB.Model(&model.Employee{}).Where("id = ?", tokenBody.EmployeeId).First(&currentEmployee)
		if result.Error != nil {
			ctx.AbortWithStatus(403)
			return
		}

		result = model.DB.Table("menu_access").Where("position_id = ?", currentEmployee.PositionId.String).First(&currentMenuAccess)
		if result.Error != nil {
			ctx.AbortWithStatus(403)
			return
		}

		err := json.Unmarshal([]byte(currentMenuAccess.Access.String), &access)
		if err != nil {
			panic(err.Error())
		}
	} else {
		var companySubscription model.Subscription
		var currentMenuAccess model.AvailablePackage

		result := model.DB.Model(&model.Subscription{}).Where("company_id = ?", tokenBody.CompanyId).First(&companySubscription)
		if result.Error != nil {
			ctx.AbortWithStatus(403)
			return
		}

		result = model.DB.Model(&model.AvailablePackage{}).Where("id = ?", companySubscription.SubscribedPackage.String).First(&currentMenuAccess)
		if result.Error != nil {
			ctx.AbortWithStatus(403)
			return
		}

		err := json.Unmarshal([]byte(currentMenuAccess.AvailableMenu.String), &access)
		if err != nil {
			panic(err.Error())
		}
	}

	if !checkAuthorization(ctx.Request.URL.Path, ctx.Request.Method, access) {
		fmt.Println("Not Authorized\n", access)
		ctx.AbortWithStatus(403)
		return
	}
	ctx.Next()
}

func checkAuthorization(url string, method string, access []AccessPermission) bool {
	var currentAccessedMenu config.MenuAccessBody

	for index, item := range config.MenuAccesses {
		if strings.Contains(url, item.Url) {
			currentAccessedMenu = item
			break
		}

		if index == len(config.MenuAccesses)-1 {
			return false
		}
	}

	if method == "GET" {
		for index, item := range access {
			if item.MainMenu == currentAccessedMenu.MainMenu {
				for _, item2 := range item.SubMenu {
					if item2 == currentAccessedMenu.SubMenu {
						return true
					}
				}
			}
			if index == len(access)-1 {
				return false
			}
		}
	} else {
		action := MethodToAction[method]

		for index, item := range access {
			if item.MainMenu == currentAccessedMenu.MainMenu {
				subMenuCheck := false
				actionCheck := false

				for _, item2 := range item.SubMenu {
					if item2 == currentAccessedMenu.SubMenu {
						subMenuCheck = true
						break
					}
				}
				if !subMenuCheck {
					fmt.Println("subMenu2")
					return false
				}

				for _, item3 := range item.Action {
					if item3 == action {
						fmt.Println("Action2")
						actionCheck = true
						break
					}
				}

				if !actionCheck {
					return false
				} else {
					return true
				}
			}

			if index == len(access)-1 {
				return false
			}
		}
	}

	return true
}
