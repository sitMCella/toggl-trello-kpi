package trello

import (
	trelloLib "github.com/adlio/trello"
	"github.com/sitMCella/toggl-trello-kpi/configuration"
	"go.uber.org/zap"
)

// TrelloClient implements the Trello Client interface.
type TrelloClient struct {
	logger        *zap.Logger
	client        *trelloLib.Client
	configuration configuration.TrelloConfiguration
}

// NewTrelloClient creates a new TrelloClient.
func NewTrelloClient(config configuration.Configuration, logger *zap.Logger, trelloClient *trelloLib.Client) *TrelloClient {
	return &TrelloClient{
		logger:        logger,
		client:        trelloClient,
		configuration: config.TrelloConfiguration,
	}
}

// GetCards retrieves all the Trello cards from the board.
func (trelloClient *TrelloClient) GetCards() ([]TrelloCardEntry, error) {
	board, err := trelloClient.client.GetBoard(trelloClient.configuration.BoardId, trelloLib.Defaults())
	if err != nil {
		return nil, err
	}
	cards, err := board.GetCards(trelloLib.Defaults())
	if err != nil {
		return nil, err
	}
	var trelloCardEntries = make([]TrelloCardEntry, len(cards))
	for i, card := range cards {
		var labels = make([]string, len(card.Labels))
		project := ""
		customer := ""
		team := ""
		cardType := ""
		for j, label := range card.Labels {
			labels[j] = label.Name
			if contains(trelloClient.configuration.LabelProjectColor, label.Color) {
				project = label.Name
			}
			if contains(trelloClient.configuration.LabelCustomerColor, label.Color) {
				customer = label.Name
			}
			if contains(trelloClient.configuration.LabelTeamColor, label.Color) {
				team = label.Name
			}
			if contains(trelloClient.configuration.LabelCardTypeColor, label.Color) {
				cardType = label.Name
			}
		}
		trelloCardEntry := TrelloCardEntry{
			Id:       card.ID,
			Name:     card.Name,
			Closed:   card.Closed,
			Labels:   labels,
			Project:  project,
			Customer: customer,
			Team:     team,
			Type:     cardType,
		}
		trelloCardEntries[i] = trelloCardEntry
	}
	return trelloCardEntries, nil
}
