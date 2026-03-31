package equity

import (
	"fmt"
	"testing"
	"time"

	"github.com/block52/poker-equity/pkg/types"
)

// =============================================================================
// Hand Evaluator Tests
// =============================================================================

func TestEvaluateHand_HighCard(t *testing.T) {
	cards, _ := CardsFromMnemonics([]string{"2H", "5D", "8C", "JS", "KH"})
	result := EvaluateHand(cards)

	if result.Rank != HighCard {
		t.Errorf("expected HighCard, got %v", result.Rank)
	}
}

func TestEvaluateHand_OnePair(t *testing.T) {
	cards, _ := CardsFromMnemonics([]string{"AS", "AH", "5D", "8C", "KH"})
	result := EvaluateHand(cards)

	if result.Rank != OnePair {
		t.Errorf("expected OnePair, got %v", result.Rank)
	}
}

func TestEvaluateHand_TwoPair(t *testing.T) {
	cards, _ := CardsFromMnemonics([]string{"AS", "AH", "KD", "KC", "5H"})
	result := EvaluateHand(cards)

	if result.Rank != TwoPair {
		t.Errorf("expected TwoPair, got %v", result.Rank)
	}
}

func TestEvaluateHand_ThreeOfAKind(t *testing.T) {
	cards, _ := CardsFromMnemonics([]string{"AS", "AH", "AD", "8C", "KH"})
	result := EvaluateHand(cards)

	if result.Rank != ThreeOfAKind {
		t.Errorf("expected ThreeOfAKind, got %v", result.Rank)
	}
}

func TestEvaluateHand_Straight(t *testing.T) {
	cards, _ := CardsFromMnemonics([]string{"5H", "6D", "7C", "8S", "9H"})
	result := EvaluateHand(cards)

	if result.Rank != Straight {
		t.Errorf("expected Straight, got %v", result.Rank)
	}
}

func TestEvaluateHand_Wheel(t *testing.T) {
	cards, _ := CardsFromMnemonics([]string{"AH", "2D", "3C", "4S", "5H"})
	result := EvaluateHand(cards)

	if result.Rank != Straight {
		t.Errorf("expected Straight (wheel), got %v", result.Rank)
	}
}

func TestEvaluateHand_Flush(t *testing.T) {
	cards, _ := CardsFromMnemonics([]string{"2H", "5H", "8H", "JH", "KH"})
	result := EvaluateHand(cards)

	if result.Rank != Flush {
		t.Errorf("expected Flush, got %v", result.Rank)
	}
}

func TestEvaluateHand_FullHouse(t *testing.T) {
	cards, _ := CardsFromMnemonics([]string{"AS", "AH", "AD", "KC", "KH"})
	result := EvaluateHand(cards)

	if result.Rank != FullHouse {
		t.Errorf("expected FullHouse, got %v", result.Rank)
	}
}

func TestEvaluateHand_FourOfAKind(t *testing.T) {
	cards, _ := CardsFromMnemonics([]string{"AS", "AH", "AD", "AC", "KH"})
	result := EvaluateHand(cards)

	if result.Rank != FourOfAKind {
		t.Errorf("expected FourOfAKind, got %v", result.Rank)
	}
}

func TestEvaluateHand_StraightFlush(t *testing.T) {
	cards, _ := CardsFromMnemonics([]string{"5H", "6H", "7H", "8H", "9H"})
	result := EvaluateHand(cards)

	if result.Rank != StraightFlush {
		t.Errorf("expected StraightFlush, got %v", result.Rank)
	}
}

func TestEvaluateHand_RoyalFlush(t *testing.T) {
	cards, _ := CardsFromMnemonics([]string{"TH", "JH", "QH", "KH", "AH"})
	result := EvaluateHand(cards)

	if result.Rank != StraightFlush {
		t.Errorf("expected StraightFlush (royal), got %v", result.Rank)
	}
}

func TestEvaluateHand_7Cards(t *testing.T) {
	cards, _ := CardsFromMnemonics([]string{"AS", "AH", "AD", "KC", "KH", "2D", "3C"})
	result := EvaluateHand(cards)

	if result.Rank != FullHouse {
		t.Errorf("expected FullHouse from 7 cards, got %v", result.Rank)
	}
}

func TestCompareHands(t *testing.T) {
	tests := []struct {
		name     string
		hand1    []string
		hand2    []string
		expected int
	}{
		{
			name:     "pair beats high card",
			hand1:    []string{"AS", "AH", "5D", "8C", "KH"},
			hand2:    []string{"2H", "5D", "8C", "JS", "KH"},
			expected: 1,
		},
		{
			name:     "higher pair wins",
			hand1:    []string{"AS", "AH", "5D", "8C", "KH"},
			hand2:    []string{"KS", "KD", "5H", "8S", "JH"},
			expected: 1,
		},
		{
			name:     "flush beats straight",
			hand1:    []string{"2H", "5H", "8H", "JH", "KH"},
			hand2:    []string{"5D", "6C", "7S", "8H", "9D"},
			expected: 1,
		},
		{
			name:     "identical hands tie",
			hand1:    []string{"AS", "KS", "QS", "JS", "9S"},
			hand2:    []string{"AH", "KH", "QH", "JH", "9H"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards1, _ := CardsFromMnemonics(tt.hand1)
			cards2, _ := CardsFromMnemonics(tt.hand2)

			result1 := EvaluateHand(cards1)
			result2 := EvaluateHand(cards2)

			cmp := CompareHands(result1, result2)
			if cmp != tt.expected {
				t.Errorf("expected %d, got %d (hand1: %v=%d, hand2: %v=%d)",
					tt.expected, cmp, result1.Rank, result1.Score, result2.Rank, result2.Score)
			}
		})
	}
}

// =============================================================================
// Equity Calculator Tests
// =============================================================================

func TestPreflopEquity_AAvsKK(t *testing.T) {
	hands := [][]string{
		{"AS", "AH"},
		{"KS", "KH"},
	}

	result, err := PreflopEquity(hands, WithSimulations(10000), WithSeed(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	aaEquity := result.Results[0].Total
	kkEquity := result.Results[1].Total

	if aaEquity < 0.75 || aaEquity > 0.88 {
		t.Errorf("AA equity %.2f%% outside expected range (75-88%%)", aaEquity*100)
	}
	if kkEquity < 0.12 || kkEquity > 0.25 {
		t.Errorf("KK equity %.2f%% outside expected range (12-25%%)", kkEquity*100)
	}

	t.Logf("AA vs KK: %.2f%% vs %.2f%% (%d simulations)",
		aaEquity*100, kkEquity*100, result.Simulations)
}

func TestPreflopEquity_CoinFlip(t *testing.T) {
	hands := [][]string{
		{"AS", "KS"},
		{"QH", "QD"},
	}

	result, err := PreflopEquity(hands, WithSimulations(10000), WithSeed(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	aksEquity := result.Results[0].Total
	qqEquity := result.Results[1].Total

	if aksEquity < 0.40 || aksEquity > 0.52 {
		t.Errorf("AKs equity %.2f%% outside expected range (40-52%%)", aksEquity*100)
	}
	if qqEquity < 0.48 || qqEquity > 0.60 {
		t.Errorf("QQ equity %.2f%% outside expected range (48-60%%)", qqEquity*100)
	}

	t.Logf("AKs vs QQ: %.2f%% vs %.2f%% (%d simulations)",
		aksEquity*100, qqEquity*100, result.Simulations)
}

func TestFlopEquity(t *testing.T) {
	hands := [][]string{
		{"AS", "KS"},
		{"QH", "QD"},
	}
	flop := []string{"KH", "7C", "2D"}

	result, err := FlopEquity(hands, flop, WithSimulations(10000), WithSeed(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	akEquity := result.Results[0].Total

	if akEquity < 0.60 {
		t.Errorf("AK equity %.2f%% should be > 60%% with top pair", akEquity*100)
	}

	t.Logf("AK vs QQ on %v: %.2f%% vs %.2f%%",
		flop, akEquity*100, result.Results[1].Total*100)
}

func TestTurnEquity(t *testing.T) {
	hands := [][]string{
		{"AS", "KS"},
		{"QH", "QD"},
	}
	board := []string{"2S", "7S", "QC", "3H"}

	result, err := TurnEquity(hands, board, WithSimulations(10000), WithSeed(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	qqEquity := result.Results[1].Total

	t.Logf("AK (flush draw) vs QQ (set) on %v: %.2f%% vs %.2f%%",
		board, result.Results[0].Total*100, qqEquity*100)

	if qqEquity < 0.70 {
		t.Errorf("QQ equity %.2f%% should be > 70%% with a set", qqEquity*100)
	}
}

func TestRiverEquity(t *testing.T) {
	hands := [][]string{
		{"AS", "KS"},
		{"QH", "QD"},
	}
	board := []string{"2S", "7S", "QC", "3H", "9S"}

	result, err := RiverEquity(hands, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	akEquity := result.Results[0].Total
	qqEquity := result.Results[1].Total

	if akEquity != 1.0 {
		t.Errorf("AK with flush should have 100%% equity, got %.2f%%", akEquity*100)
	}
	if qqEquity != 0.0 {
		t.Errorf("QQ with set should have 0%% equity vs flush, got %.2f%%", qqEquity*100)
	}

	t.Logf("AK (flush) vs QQ (set) on river: %.2f%% vs %.2f%%",
		akEquity*100, qqEquity*100)
}

func TestMultipleHands(t *testing.T) {
	hands := [][]string{
		{"AS", "AH"},
		{"KS", "KH"},
		{"QS", "QH"},
		{"JC", "TC"},
	}

	result, err := PreflopEquity(hands, WithSimulations(10000), WithSeed(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	totalEquity := 0.0
	for _, r := range result.Results {
		totalEquity += r.Total
	}

	if totalEquity < 0.98 || totalEquity > 1.02 {
		t.Errorf("total equity %.4f should be ~1.0", totalEquity)
	}

	t.Logf("4-way: AA=%.2f%%, KK=%.2f%%, QQ=%.2f%%, JTs=%.2f%%",
		result.Results[0].Total*100,
		result.Results[1].Total*100,
		result.Results[2].Total*100,
		result.Results[3].Total*100)
}

func TestMaxHands(t *testing.T) {
	hands := [][]string{
		{"AS", "AH"}, {"KS", "KH"}, {"QS", "QH"},
		{"JS", "JH"}, {"TS", "TH"}, {"9S", "9H"},
		{"8S", "8H"}, {"7C", "7D"}, {"6C", "6D"},
	}

	result, err := PreflopEquity(hands, WithSimulations(5000), WithSeed(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	totalEquity := 0.0
	for _, r := range result.Results {
		totalEquity += r.Total
	}

	if totalEquity < 0.98 || totalEquity > 1.02 {
		t.Errorf("total equity %.4f should be ~1.0", totalEquity)
	}

	t.Logf("9-way pot: AA=%.2f%%", result.Results[0].Total*100)
}

func TestDuplicateCardError(t *testing.T) {
	hands := [][]string{
		{"AS", "AH"},
		{"AS", "KH"},
	}

	_, err := PreflopEquity(hands)
	if err == nil {
		t.Error("expected error for duplicate cards")
	}
}

func TestInvalidHandSize(t *testing.T) {
	hands := [][]string{
		{"AS", "AH", "KH"},
		{"KS", "KH"},
	}

	_, err := PreflopEquity(hands)
	if err == nil {
		t.Error("expected error for invalid hand size")
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkEvaluateHand5Cards(b *testing.B) {
	cards, _ := CardsFromMnemonics([]string{"AS", "KS", "QS", "JS", "TS"})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EvaluateHand(cards)
	}
}

func BenchmarkEvaluateHand7Cards(b *testing.B) {
	cards, _ := CardsFromMnemonics([]string{"AS", "KS", "QS", "JS", "TS", "2H", "3D"})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EvaluateHand(cards)
	}
}

func BenchmarkEvaluateHandFast7Cards(b *testing.B) {
	cards, _ := CardsFromMnemonics([]string{"AS", "KS", "QS", "JS", "TS", "2H", "3D"})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EvaluateHandFast(cards)
	}
}

func BenchmarkPreflopEquity2Hands(b *testing.B) {
	hands := [][]string{{"AS", "AH"}, {"KS", "KH"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PreflopEquity(hands, WithSimulations(1000))
	}
}

func BenchmarkPreflopEquity9Hands(b *testing.B) {
	hands := [][]string{
		{"AS", "AH"}, {"KS", "KH"}, {"QS", "QH"},
		{"JS", "JH"}, {"TS", "TH"}, {"9S", "9H"},
		{"8S", "8H"}, {"7C", "7D"}, {"6C", "6D"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PreflopEquity(hands, WithSimulations(1000))
	}
}

// =============================================================================
// Speed Diagnostics
// =============================================================================

func TestSpeedDiagnostics(t *testing.T) {
	fmt.Println("")
	fmt.Println("=== EQUITY CALCULATOR SPEED DIAGNOSTICS ===")

	cards5, _ := CardsFromMnemonics([]string{"AS", "KS", "QS", "JS", "TS"})
	cards7, _ := CardsFromMnemonics([]string{"AS", "KS", "QS", "JS", "TS", "2H", "3D"})

	iterations := 100000

	start := time.Now()
	for i := 0; i < iterations; i++ {
		EvaluateHand(cards5)
	}
	duration := time.Since(start)
	fmt.Printf("  5-card evaluation: %d evals in %v (%.0f evals/sec)\n",
		iterations, duration, float64(iterations)/duration.Seconds())

	start = time.Now()
	for i := 0; i < iterations; i++ {
		EvaluateHand(cards7)
	}
	duration = time.Since(start)
	fmt.Printf("  7-card evaluation: %d evals in %v (%.0f evals/sec)\n",
		iterations, duration, float64(iterations)/duration.Seconds())

	start = time.Now()
	for i := 0; i < iterations; i++ {
		EvaluateHandFast(cards7)
	}
	duration = time.Since(start)
	fmt.Printf("  7-card fast eval:  %d evals in %v (%.0f evals/sec)\n",
		iterations, duration, float64(iterations)/duration.Seconds())

	fmt.Println("\nEquity Calculation Performance:")
	hands2 := [][]string{{"AS", "AH"}, {"KS", "KH"}}

	for _, sims := range []int{1000, 10000, 50000} {
		result, _ := PreflopEquity(hands2, WithSimulations(sims))
		fmt.Printf("  %d sims, 2-way preflop: %v (%.0f hands/sec)\n",
			sims, result.Duration, result.HandsPerSec)
	}

	fmt.Println("=== END SPEED DIAGNOSTICS ===")
	fmt.Println("")
}

// =============================================================================
// Accuracy Tests
// =============================================================================

func TestAccuracyVsKnownEquities(t *testing.T) {
	tests := []struct {
		name           string
		hand1          []string
		hand2          []string
		expectedEquity float64
		tolerance      float64
	}{
		{"AA vs KK", []string{"AS", "AH"}, []string{"KS", "KH"}, 0.82, 0.03},
		{"AA vs 72o", []string{"AS", "AH"}, []string{"7D", "2C"}, 0.88, 0.03},
		{"AKs vs QQ", []string{"AS", "KS"}, []string{"QH", "QD"}, 0.46, 0.03},
		{"KK vs AKo", []string{"KS", "KH"}, []string{"AD", "KC"}, 0.70, 0.03},
		{"22 vs AKo", []string{"2S", "2H"}, []string{"AD", "KC"}, 0.52, 0.03},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hands := [][]string{tt.hand1, tt.hand2}
			result, err := PreflopEquity(hands, WithSimulations(50000), WithSeed(42))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			actualEquity := result.Results[0].Total
			diff := actualEquity - tt.expectedEquity
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.tolerance {
				t.Errorf("%s: expected %.2f%%, got %.2f%% (diff: %.2f%%)",
					tt.name, tt.expectedEquity*100, actualEquity*100, diff*100)
			}

			t.Logf("%s: %.2f%% vs %.2f%% (expected ~%.0f%%)",
				tt.name, actualEquity*100, result.Results[1].Total*100, tt.expectedEquity*100)
		})
	}
}

// =============================================================================
// Fast Evaluator Tests
// =============================================================================

func TestFastEvaluator_AllHandTypes(t *testing.T) {
	tests := []struct {
		name     string
		cards    []string
		category uint8
	}{
		{"High Card", []string{"2H", "5D", "8C", "JS", "KH", "3C", "7D"}, 0},
		{"One Pair", []string{"AS", "AH", "5D", "8C", "KH", "2D", "3C"}, 1},
		{"Two Pair", []string{"AS", "AH", "KD", "KC", "5H", "2D", "3C"}, 2},
		{"Three of a Kind", []string{"AS", "AH", "AD", "8C", "KH", "2D", "3C"}, 3},
		{"Straight", []string{"5H", "6D", "7C", "8S", "9H", "2D", "3C"}, 4},
		{"Flush", []string{"2H", "5H", "8H", "JH", "KH", "3D", "4C"}, 5},
		{"Full House", []string{"AS", "AH", "AD", "KC", "KH", "2D", "3C"}, 6},
		{"Four of a Kind", []string{"AS", "AH", "AD", "AC", "KH", "2D", "3C"}, 7},
		{"Straight Flush", []string{"5H", "6H", "7H", "8H", "9H", "2D", "3C"}, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards, _ := CardsFromMnemonics(tt.cards)
			result := EvaluateHandFast(cards)

			if result.Category != tt.category {
				t.Errorf("expected category %d (%s), got %d", tt.category, tt.name, result.Category)
			}
		})
	}
}

func TestFastEvaluator_Ordering(t *testing.T) {
	hands := [][]string{
		{"2H", "5D", "8C", "JS", "KH", "3C", "7D"}, // High card
		{"AS", "AH", "5D", "8C", "KH", "2D", "3C"}, // Pair
		{"AS", "AH", "KD", "KC", "5H", "2D", "3C"}, // Two pair
		{"AS", "AH", "AD", "8C", "KH", "2D", "3C"}, // Trips
		{"5H", "6D", "7C", "8S", "9H", "2D", "3C"}, // Straight
		{"2H", "5H", "8H", "JH", "KH", "3D", "4C"}, // Flush
		{"AS", "AH", "AD", "KC", "KH", "2D", "3C"}, // Full house
		{"AS", "AH", "AD", "AC", "KH", "2D", "3C"}, // Quads
		{"5H", "6H", "7H", "8H", "9H", "2D", "3C"}, // Straight flush
	}

	var lastScore uint32
	for i, hand := range hands {
		cards, _ := CardsFromMnemonics(hand)
		result := EvaluateHandFast(cards)

		if result.Score <= lastScore && i > 0 {
			t.Errorf("hand %d (%v) score %d should be > hand %d score %d",
				i, hand, result.Score, i-1, lastScore)
		}
		lastScore = result.Score
	}
}

// =============================================================================
// Card Parsing Tests
// =============================================================================

func TestCardsFromMnemonics(t *testing.T) {
	mnemonics := []string{"AS", "KH", "QD", "JC", "TH"}
	cards, err := CardsFromMnemonics(mnemonics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []struct {
		rank int
		suit types.Suit
	}{
		{1, types.SuitSpades},
		{13, types.SuitHearts},
		{12, types.SuitDiamonds},
		{11, types.SuitClubs},
		{10, types.SuitHearts},
	}

	for i, exp := range expected {
		if cards[i].Rank != exp.rank || cards[i].Suit != exp.suit {
			t.Errorf("card %d: expected rank=%d suit=%d, got rank=%d suit=%d",
				i, exp.rank, exp.suit, cards[i].Rank, cards[i].Suit)
		}
	}
}

func TestCardsFromMnemonics_InvalidCard(t *testing.T) {
	mnemonics := []string{"AS", "XX"}
	_, err := CardsFromMnemonics(mnemonics)
	if err == nil {
		t.Error("expected error for invalid card mnemonic")
	}
}
