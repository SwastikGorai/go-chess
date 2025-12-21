package api

const (
	resultOngoing   = "ongoing"
	resultCheckmate = "checkmate"
	resultStalemate = "stalemate"
	resultDraw      = "draw"
	resultResigned  = "resigned"
)

const (
	endedByCheckmate            = "checkmate"
	endedByStalemate            = "stalemate"
	endedByResignation          = "resignation"
	endedByDrawAgreement        = "draw_agreement"
	endedByDrawClaim            = "draw_claim"
	endedByInsufficientMaterial = "insufficient_material"
	endedByFiftyMove            = "fifty_move"
)

type Status struct {
	Result  string
	Winner  string
	EndedBy string
	Flags   Flags
}

func computeStatus(game *Game) Status {
	board := game.Board
	flags := Flags{
		InCheck: board.InCheck(board.Turn()),
	}

	if game.Result != "" {
		applyStoredStatus(&flags, game.Result, game.EndedBy)
		return Status{
			Result:  game.Result,
			Winner:  game.Winner,
			EndedBy: game.EndedBy,
			Flags:   flags,
		}
	}

	if board.IsCheckmate(board.Turn()) {
		flags.Checkmate = true
		flags.InCheck = true
		return Status{
			Result:  resultCheckmate,
			Winner:  board.Turn().Opposite().String(),
			EndedBy: endedByCheckmate,
			Flags:   flags,
		}
	}

	if board.IsStalemate(board.Turn()) {
		flags.Stalemate = true
		flags.Draw = true
		return Status{
			Result:  resultStalemate,
			Winner:  "none",
			EndedBy: endedByStalemate,
			Flags:   flags,
		}
	}

	if board.IsInsufficientMaterial() {
		flags.Draw = true
		flags.DrawReason = endedByInsufficientMaterial
		return Status{
			Result:  resultDraw,
			Winner:  "none",
			EndedBy: endedByInsufficientMaterial,
			Flags:   flags,
		}
	}

	if board.CanClaimFiftyMoveDraw() {
		flags.DrawClaimable = true
		flags.DrawReason = endedByFiftyMove
	}

	return Status{
		Result: resultOngoing,
		Flags:  flags,
	}
}

func applyStoredStatus(flags *Flags, result string, endedBy string) {
	switch result {
	case resultResigned:
		flags.Draw = false
	case resultDraw:
		flags.Draw = true
		if endedBy == endedByInsufficientMaterial || endedBy == endedByFiftyMove {
			flags.DrawReason = endedBy
		}
	case resultStalemate:
		flags.Stalemate = true
		flags.Draw = true
	case resultCheckmate:
		flags.Checkmate = true
		flags.InCheck = true
	}
}
