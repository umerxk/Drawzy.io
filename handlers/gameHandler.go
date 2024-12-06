package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"skribbl-clone/db"
	"skribbl-clone/utils"

	"github.com/gofiber/fiber/v2"
)

type ApiPayload struct {
	Params struct {
		PageNumber int `json:"pageNumber"`
		PageSize   int `json:"pageSize"`
		Sorting    struct {
			Key  string `json:"key"`
			Type int    `json:"type"`
		} `json:"sorting"`
		Search        []string `json:"search"`
		HideTrialTeam bool     `json:"hideTrialTeam"`
		GSearch       string   `json:"gSearch"`
		ClientId      string   `json:"clientId"`
		Tier          []string `json:"tier"`
	} `json:"params"`
	UserId string `json:"userId"`
}

type ApiResponse struct {
	Teams      []map[string]interface{} `json:"teams"`      // Extract "teams"
	Tiers      []map[string]interface{} `json:"tiers"`      // Optional, but not used here
	TotalTeams int                      `json:"totalTeams"` // Optional, but not used here
}

func fetchTeamsData() ([]map[string]interface{}, error) {
	// API URL
	url := "https://develop-listener.arthur-dev.digital/teams/allTeams"

	// JSON payload
	payload := ApiPayload{
		UserId: "618bc99b7690879b6af592de",
	}
	payload.Params.PageNumber = 1
	payload.Params.PageSize = 75
	payload.Params.Sorting.Key = "_id"
	payload.Params.Sorting.Type = 1
	payload.Params.Search = []string{}
	payload.Params.HideTrialTeam = false
	payload.Params.GSearch = ""
	payload.Params.ClientId = ""
	payload.Params.Tier = []string{}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}
	log.Printf("Payload: %s", jsonData)
	// Create a new request
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers, including the x-access-token
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-access-token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzQwMDI2Njk4MTEsImRhdGEiOnsiZW1haWwiOiJ1bWVyLnNoZXJhekBhcnRodXIuZGlnaXRhbCIsInVzZXJJZCI6IjYxOGJjOTliNzY5MDg3OWI2YWY1OTJkZSIsInVzZXJEZXZpY2VJZCI6ImViZWRlZGUyLTg2OWEtNDgxYi04MTUzLWVhMDQ3MThjYTUzNyJ9LCJpYXQiOjE3MzI3MDY2Njk4MTF9.EJ37HaAIV_HDtm4uYjDl2bq1rEdefiyXkW81W2dH_0E") // Replace with the actual token

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var apiResponse ApiResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return apiResponse.Teams, nil
}

func GetGames(c *fiber.Ctx) error {

	data, err := fetchTeamsData()
	// log.Println(string(data))

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error fetching data: %v", err))
	}

	// type Team struct {
	// 	TeamID   int    `json:"team_id"`
	// 	TeamName string `json:"team_name"`
	// 	TeamKey  string `json:"team_key"`
	// }
	// file, err := os.Open("teamdata.json")
	// if err != nil {
	// 	log.Fatalf("Failed to open file: %v", err)
	// 	return c.Status(500).SendString("Failed to read the file")
	// }
	// defer file.Close()

	// var teams []Team
	// if err := json.NewDecoder(file).Decode(&teams); err != nil {
	// 	log.Fatalf("Failed to decode JSON: %v", err)
	// 	return c.Status(500).SendString("Failed to parse JSON")
	// }

	// for _, team := range teams {
	// 	fmt.Printf("Team ID: %d, Team Name: %s, Team Key: %s\n", team.TeamID, team.TeamName, team.TeamKey)
	// }
	return c.JSON(fiber.Map{"status": "success", "message": "All games", "data": (data)})
}

type TeamMember struct {
	User            string `json:"user"`
	UserRole        string `json:"userRole"`
	UserTeamRole    string `json:"userTeamRole"`
	DefaultUserRole bool   `json:"defaultUserRole"`
}

type Platforms struct {
	Name    int `json:"name"`
	Version int `json:"version"`
}

type TeamDlcs struct {
	DlcId     string      `json:"dlcId"`
	Platforms []Platforms `json:"platforms"`
}

type Request struct {
	Id                   string       `json:"_id"`
	TeamName             string       `json:"teamName"`
	TeamKey              string       `json:"teamKey"`
	CreatedBy            string       `json:"createdBy"`
	Tier                 string       `json:"tier"`
	TierPrice            float64      `json:"tierPrice"`
	Organization         string       `json:"organization"`
	IsDefaultTeam        bool         `json:"isDefaultTeam"`
	Subscription         string       `json:"subscription"`
	CanceledAt           int64        `json:"canceledAt"`
	CreatedAt            string       `json:"createdAt"`
	TeamMembers          []TeamMember `json:"teamMembers"`
	TeamDlcs             []TeamDlcs   `json:"teamDlcs"`
	InvitedMembers       []string     `json:"invitedMembers"`
	SavedRooms           []string     `json:"savedRooms"`
	CanceledSubscription []string     `json:"canceledSubscription"`
}

func CreateGame(c *fiber.Ctx) error {
	var teamId int
	var mongodbId string
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Invalid JSON", err)
	}
	tx, err := db.DB.Begin()

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to start database transaction", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			log.Println("Transaction rolled back due to error")
		}
	}()

	teamsQuery := `
    INSERT INTO teams (
        mongodb_id, 
        team_name, 
        team_key, 
        created_by, 
        tier, 
        tier_price, 
        organization, 
        is_default_team, 
        subscription, 
        canceled_at
    ) 
    VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
    ) RETURNING id, mongodb_id`

	err = tx.QueryRow(teamsQuery, req.Id, req.TeamName, req.TeamKey, req.CreatedBy, req.Tier, req.TierPrice, req.Organization, req.IsDefaultTeam, req.Subscription, req.CanceledAt).Scan(&teamId, &mongodbId)

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to insert team data", err)
	}

	teamMebersQuery := `
    INSERT INTO team_members (
        team_id, 
        "user", 
        user_role, 
        user_team_role, 
        default_user_role
    ) 
    VALUES ($1, $2, $3, $4, $5)`

	for _, member := range req.TeamMembers {
		_, err := tx.Exec(teamMebersQuery,
			teamId,
			member.User,
			member.UserRole,
			member.UserTeamRole,
			member.DefaultUserRole,
		)

		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to insert team members", err)
		}
	}

	teamDlcsQuery := `
    INSERT INTO team_dlcs (
        team_id, 
        dlc_id
    ) 
    VALUES ($1, $2) RETURNING id`
	var teamDlcId int
	for _, dlc := range req.TeamDlcs {
		err := tx.QueryRow(teamDlcsQuery, teamId, dlc.DlcId).Scan(&teamDlcId)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to insert team DLC data", err)
		}
		for _, platform := range dlc.Platforms {
			_, err := tx.Exec(`INSERT INTO dlc_platforms (dlc_id, platform_name, platform_version) VALUES ($1, $2, $3)`,
				teamDlcId, platform.Name, platform.Version)
			if err != nil {
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to insert DLC platform data", err)
			}
		}
	}

	invitedMembersQuery := `
    INSERT INTO invited_members (
		team_id,
        team_member_id
    ) 
    VALUES ($1, $2)`

	for _, invitedMember := range req.InvitedMembers {
		_, err := tx.Exec(invitedMembersQuery,
			teamId,
			invitedMember,
		)

		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to insert invited members", err)
		}
	}

	savedRoomsQuery := `
    INSERT INTO saved_rooms (
		team_id,
        room_id
    ) 
    VALUES ($1, $2)`

	for _, room := range req.SavedRooms {
		_, err := tx.Exec(savedRoomsQuery,
			teamId,
			room,
		)

		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to insert saved rooms", err)
		}
	}

	subscriptionsQuery := `
    INSERT INTO canceled_subscriptions (
		team_id,
        subscription_id
    ) 
    VALUES ($1, $2)`

	for _, subscriptionId := range req.CanceledSubscription {
		_, err := tx.Exec(subscriptionsQuery,
			teamId,
			subscriptionId,
		)

		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to insert canceledSubscription", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to commit transaction", err)
	}
	log.Println("Transaction committed successfully!")
	return c.JSON(fiber.Map{
		"message": "Data inserted successfully",
		"team_id": teamId,
	})
}

func GetGameByID(c *fiber.Ctx) error {
	id := c.Params("id")
	// Fetch game by ID
	return c.JSON(fiber.Map{"status": "success", "message": "Game found", "id": id})
}
