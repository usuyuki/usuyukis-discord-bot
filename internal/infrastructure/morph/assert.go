package morph

import haikuUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/haiku"

// コンパイル時にportを満たしていることを保証する
var _ haikuUC.MorphAnalyzer = (*KagomeAnalyzer)(nil)
