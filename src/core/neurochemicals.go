package core

// Este arquivo conteria a lógica mais detalhada para Cortisol e Dopamina se ela
// fosse muito extensa e não coubesse bem em network.go.
// No entanto, a maior parte da lógica (produção, decaimento, efeitos diretos)
// já foi integrada em network.go e neuron.go para o MVP.


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
