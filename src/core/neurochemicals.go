package core

// Este arquivo conteria a lógica mais detalhada para Cortisol e Dopamina se ela
// fosse muito extensa e não coubesse bem em network.go.
// No entanto, a maior parte da lógica (produção, decaimento, efeitos diretos)
// já foi integrada em network.go e neuron.go para o MVP.

// Por exemplo, se a modulação do limiar pelo cortisol fosse uma curva complexa,
// essa função poderia residir aqui.
// func CalculateCortisolEffectOnThreshold(cortisolLevel float64, baseEffect float64) float64 {
//     //  "diminui o limiar de disparo dos neurônios inicialmente,
//     //   mas ao atingir um pico, começa a reduzir o limiar [efeito de diminuição],
//     //   diminuindo também a sinapogênese."
//
//     // Interpretação:
//     // Nível Baixo (0 a ~0.3): Pequena diminuição do limiar.
//     // Nível Médio (~0.3 a ~0.8): Maior diminuição do limiar (pico do efeito benéfico).
//     // Nível Alto (~0.8 a ~1.5): Diminuição do limiar menos pronunciada, voltando ao normal.
//     // Nível Muito Alto (> ~1.5): Aumento do limiar (efeito deletério).
//
//     if cortisolLevel < 0.3 {
//         return -cortisolLevel * baseEffect // ex: -0.0 para -0.03 (se baseEffect=0.1)
//     } else if cortisolLevel < 0.8 {
//         // (cortisolLevel-0.3)* ALGO_PARA_SUBIR + (0.8-cortisolLevel)*ALGO_PARA_DESCER
//         // Função tipo sino/parábola invertida.
//         // Maxima redução em ~0.55
//         normalizedPeak := (cortisolLevel - 0.3) / (0.8 - 0.3) // 0 a 1
//         return -(0.3 + (0.5 * (1.0 - math.Abs(normalizedPeak-0.5)*2.0))) * baseEffect
//     } else if cortisolLevel < 1.5 {
//         // Efeito de diminuição do limiar reduz
//         return -( (1.5 - cortisolLevel) / (1.5 - 0.8) * 0.4 ) * baseEffect // De -0.4*base a 0
//     } else {
//         // Aumento do limiar
//         return (cortisolLevel - 1.5) * baseEffect * 0.5 // Aumento mais lento
//     }
// }

// A lógica de produção, decaimento e os efeitos principais de Cortisol e Dopamina
// estão atualmente implementados em:
// - `Network.SimulateCycle()`:
//    - Produção de cortisol pela glândula.
//    - Produção de dopamina por neurônios dopaminérgicos que disparam.
//    - Decaimento de ambos os neuroquímicos.
//    - Aplicação dos efeitos no limiar dos neurônios.
//    - Modulação da taxa de sinaptogênese (em `ApplySynaptogenesis` via `net.Config` e níveis atuais).
// - `Neuron.AdjustFiringThreshold()`: Aplica o ajuste calculado ao limiar do neurônio.

// Manter este arquivo como um placeholder caso precisemos de funções auxiliares
// mais complexas para neuroquímicos no futuro, ou para refatorar a lógica
// de `network.go` se ela se tornar muito grande.

// Por exemplo, a lógica de como a "recompensa" ou "punição" (do aprendizado por reforço)
// afeta os níveis de cortisol/dopamina poderia ser colocada aqui.

// RewardSignal modifica os níveis de neuroquímicos baseado num sinal de recompensa.
// `reward` > 0 para recompensa, < 0 para punição.
func ApplyRewardSignal(net *Network, reward float64) {
	// Exemplo simples:
	// Recompensa aumenta dopamina e/ou reduz cortisol.
	// Punição aumenta cortisol e/ou reduz dopamina.
	if reward > 0 { // Recompensa
		net.DopamineLevel += reward * 0.2 // Aumento proporcional à recompensa
		// net.CortisolLevel -= reward * 0.1 // Redução proporcional
		// if net.CortisolLevel < 0 {
		// 	net.CortisolLevel = 0
		// }
	} else if reward < 0 { // Punição (reward é negativo)
		net.CortisolLevel -= reward * 0.15 // Aumenta cortisol (reward é negativo, ex: -(-1)*0.1 = +0.1)
		// net.DopamineLevel += reward * 0.05 // Reduz dopamina (reward é negativo)
		// if net.DopamineLevel < 0 {
		// 	net.DopamineLevel = 0
		// }
	}
	// Clampar os níveis se necessário
	if net.DopamineLevel < 0 { net.DopamineLevel = 0 }
	if net.DopamineLevel > 5.0 { net.DopamineLevel = 5.0 } // Um limite superior arbitrário
	if net.CortisolLevel < 0 { net.CortisolLevel = 0 }
	if net.CortisolLevel > 5.0 { net.CortisolLevel = 5.0 } // Um limite superior arbitrário
}
