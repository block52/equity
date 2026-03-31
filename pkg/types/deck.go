package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// Suit represents a card suit
type Suit int

const (
	SuitClubs    Suit = 1
	SuitDiamonds Suit = 2
	SuitHearts   Suit = 3
	SuitSpades   Suit = 4
)

// Card represents a playing card
type Card struct {
	Suit     Suit   `json:"suit"`
	Rank     int    `json:"rank"`
	Value    int    `json:"value"`
	Mnemonic string `json:"mnemonic"`
}

// Deck represents a deck of playing cards
type Deck struct {
	cards    []Card
	Hash     string
	SeedHash string
	top      int
}

// NewDeck creates a new deck from a string representation or initializes a standard 52-card deck
func NewDeck(deckStr string) (*Deck, error) {
	deck := &Deck{
		cards:    make([]Card, 0),
		Hash:     "",
		SeedHash: "",
		top:      0,
	}

	if deckStr != "" {
		mnemonics := strings.Split(deckStr, "-")
		if len(mnemonics) != 52 {
			return nil, fmt.Errorf("deck must contain 52 cards, got %d", len(mnemonics))
		}

		for i, mnemonic := range mnemonics {
			if strings.HasPrefix(mnemonic, "[") && strings.HasSuffix(mnemonic, "]") {
				mnemonic = mnemonic[1 : len(mnemonic)-1]
				deck.top = i
			}

			card, err := CardFromString(mnemonic)
			if err != nil {
				return nil, fmt.Errorf("invalid card at position %d: %w", i, err)
			}

			deck.cards = append(deck.cards, card)
		}
	} else {
		deck.initStandard52()
	}

	deck.createHash()
	return deck, nil
}

// Shuffle shuffles the deck using Fisher-Yates algorithm with a seed
func (d *Deck) Shuffle(seed []int) {
	seedStr := fmt.Sprintf("%v", seed)
	hash := sha256.Sum256([]byte(seedStr))
	d.SeedHash = hex.EncodeToString(hash[:])

	for i := len(d.cards) - 1; i > 0; i-- {
		j := seed[i] % (i + 1)
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	}

	d.createHash()
}

// GetNext returns the next card from the top of the deck
func (d *Deck) GetNext() Card {
	card := d.cards[d.top]
	d.top++
	return card
}

// Deal deals the specified number of cards from the deck
func (d *Deck) Deal(amount int) []Card {
	cards := make([]Card, amount)
	for i := 0; i < amount; i++ {
		cards[i] = d.GetNext()
	}
	return cards
}

// ToString serializes the deck to a string representation
func (d *Deck) ToString() string {
	mnemonics := make([]string, len(d.cards))

	for i, card := range d.cards {
		if i == d.top {
			mnemonics[i] = fmt.Sprintf("[%s]", card.Mnemonic)
		} else {
			mnemonics[i] = card.Mnemonic
		}
	}

	return strings.Join(mnemonics, "-")
}

// GetCardMnemonic returns the mnemonic representation of a card
func GetCardMnemonic(suit Suit, rank int) string {
	var rankStr string
	switch rank {
	case 1:
		rankStr = "A"
	case 10:
		rankStr = "T"
	case 11:
		rankStr = "J"
	case 12:
		rankStr = "Q"
	case 13:
		rankStr = "K"
	default:
		rankStr = strconv.Itoa(rank)
	}

	var suitStr string
	switch suit {
	case SuitClubs:
		suitStr = "C"
	case SuitDiamonds:
		suitStr = "D"
	case SuitHearts:
		suitStr = "H"
	case SuitSpades:
		suitStr = "S"
	}

	return rankStr + suitStr
}

// CardFromString parses a card from its mnemonic representation
func CardFromString(mnemonic string) (Card, error) {
	if len(mnemonic) < 2 {
		return Card{}, fmt.Errorf("invalid card mnemonic: %s", mnemonic)
	}

	rankStr := strings.ToUpper(mnemonic[:len(mnemonic)-1])
	suitChar := strings.ToUpper(string(mnemonic[len(mnemonic)-1]))

	var rank int
	switch rankStr {
	case "A":
		rank = 1
	case "T":
		rank = 10
	case "J":
		rank = 11
	case "Q":
		rank = 12
	case "K":
		rank = 13
	default:
		var err error
		rank, err = strconv.Atoi(rankStr)
		if err != nil {
			return Card{}, fmt.Errorf("invalid rank: %s", rankStr)
		}
	}

	var suit Suit
	switch suitChar {
	case "C":
		suit = SuitClubs
	case "D":
		suit = SuitDiamonds
	case "H":
		suit = SuitHearts
	case "S":
		suit = SuitSpades
	default:
		return Card{}, fmt.Errorf("invalid suit character: %s", suitChar)
	}

	value := 13*(int(suit)-1) + (rank - 1)

	return Card{
		Suit:     suit,
		Rank:     rank,
		Value:    value,
		Mnemonic: mnemonic,
	}, nil
}

func (d *Deck) createHash() {
	mnemonics := make([]string, len(d.cards))
	for i, card := range d.cards {
		mnemonics[i] = card.Mnemonic
	}

	cardsStr := strings.Join(mnemonics, "-")
	hash := sha256.Sum256([]byte(cardsStr))
	d.Hash = hex.EncodeToString(hash[:])
}

func (d *Deck) initStandard52() {
	d.cards = make([]Card, 0, 52)

	for suit := SuitClubs; suit <= SuitSpades; suit++ {
		for rank := 1; rank <= 13; rank++ {
			mnemonic := GetCardMnemonic(suit, rank)
			value := 13*(int(suit)-1) + (rank - 1)

			d.cards = append(d.cards, Card{
				Suit:     suit,
				Rank:     rank,
				Value:    value,
				Mnemonic: mnemonic,
			})
		}
	}

	d.createHash()
}
