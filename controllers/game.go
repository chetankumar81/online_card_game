package controllers

import (
	"demo_app/models"
	"demo_app/util"
	"encoding/json"
	"kbyp-common-libs/utility/log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/spf13/cast"
)

type StartGameRequest struct {
	User1 string `json:"user1"`
	User2 string `json:"user2"`
}

type PickCardRequest struct {
	GameId     string `json:"gameId"`
	User       string `json:"user"`
	Card       string `json:"card"`
	PickedTime string `json:"pickedTime"`
}

//StartGame ...
func StartGame(request events.APIGatewayProxyRequest) (response string) {
	responseJson := util.ResponseJSON{}
	responseJson.Code = 454
	responseJson.Model = "Invalid Params for starting game"
	response = util.GetResponseJSONInString(responseJson)

	var startGameRequest StartGameRequest
	err := json.Unmarshal([]byte(request.Body), &startGameRequest)
	if err != nil {
		log.Print("Error: ", err)
		return
	}

	user1, err := models.GetUserByuserName(startGameRequest.User1)
	if err != nil {
		log.Println("Error in GetUserByuserName", err)
		return
	}
	user2, err := models.GetUserByuserName(startGameRequest.User2)
	if err != nil {
		log.Println("Error in GetUserByuserName", err)
		return
	}

	if user1 == nil || user2 == nil {
		responseJson.Model = "Invalid userIds"
		response = util.GetResponseJSONInString(responseJson)
		return
	}

	isUsersAlreadyInGame, err := models.CheckUsersAlreadyInGame(user1.Id, user2.Id)
	if err != nil || isUsersAlreadyInGame {
		responseJson.Code = 455
		responseJson.Model = "User1 / user2 already in game"
		response = util.GetResponseJSONInString(responseJson)
		return
	}

	game := &models.Game{}
	game.User1 = user1
	game.User2 = user2
	game.Status = 1
	game.Timer = "30"

	gameId, err := models.AddGame(game)
	if err != nil {
		log.Println("Error in AddGame", err)
		return
	}
	game.Id = cast.ToInt(gameId)

	responseModel := make(map[string]interface{})
	responseModel["gameDetails"] = game

	responseJson.Model = responseModel
	responseJson.Code = 200
	response = util.GetResponseJSONInString(responseJson)
	return
}

//PickCard ...
func PickCard(request events.APIGatewayProxyRequest) (response string) {
	responseJson := util.ResponseJSON{}
	responseJson.Code = 400
	responseJson.Model = "Error in picking card"
	response = util.GetResponseJSONInString(responseJson)

	var pickCardRequest PickCardRequest
	err := json.Unmarshal([]byte(request.Body), &pickCardRequest)
	if err != nil {
		log.Print("Error: ", err)
		return
	}

	user, err := models.GetUserByuserName(pickCardRequest.User)
	if err != nil {
		log.Println("Error in GetUserByuserName", err)
		responseJson.Model = "Invalid User"
		response = util.GetResponseJSONInString(responseJson)
		return
	}
	game, err := models.GetGameById(cast.ToInt(pickCardRequest.GameId))
	if err != nil {
		log.Println("Error in GetGameById", err)
		responseJson.Model = "Invalid GameId"
		response = util.GetResponseJSONInString(responseJson)
		return
	}
	card, err := models.GetCardMapById(cast.ToInt(pickCardRequest.Card))
	if err != nil {
		log.Println("Error in GetCardMapById", err)
		responseJson.Model = "Invalid card"
		response = util.GetResponseJSONInString(responseJson)
		return
	}

	flag := false
	last3Cards, err := models.GetLast3CardValue(game.Id, user.Id)
	if err == nil && card.CardVal > last3Cards[2] && last3Cards[0] > last3Cards[1] && last3Cards[1] > last3Cards[2] {
		flag = true
	}

	cards := &models.Cards{}
	cards.GameId = game
	cards.UserId = user
	cards.Card = card
	cards.PickedTime = cast.ToTime(pickCardRequest.PickedTime)

	_, err = models.AddCards(cards)

	responseModel := make(map[string]interface{})
	if flag {
		game.Status = 0
		game.Result = user.Id
		game.Ended = time.Now()

		err = models.UpdateGameById(game)
		responseModel["winner"] = user.UserName

	}

	responseJson.Msg = "Success"
	responseJson.Model = responseModel
	responseJson.Code = 200
	response = util.GetResponseJSONInString(responseJson)
	return

}

//GetGameDetails ...
func GetGameDetails(request events.APIGatewayProxyRequest) (response string) {
	responseJson := util.ResponseJSON{}
	responseJson.Code = 404
	responseJson.Model = "Error Game Not Found"
	response = util.GetResponseJSONInString(responseJson)

	gameId, ok := request.QueryStringParameters["gameId"]
	if !ok {
		log.Println("gameId id not present")
		return
	}

	game, err := models.GetGameById(cast.ToInt(gameId))
	if err != nil {
		log.Print("Error in GetGameById", err)
		return
	}

	cards, err := models.GetCardPicksByGameId(game.Id)
	if err != nil {
		log.Println("Error in GetCardPicksByGameId", err)
		return
	}

	responseModel := make(map[string]interface{})
	responseModel["gameDetails"] = game
	responseModel["cardDetails"] = cards

	responseJson.Model = responseModel
	responseJson.Msg = "Success"
	responseJson.Code = 200
	response = util.GetResponseJSONInString(responseJson)
	return
}
