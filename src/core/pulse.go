package core

// Pulse representa um sinal elétrico propagando pela rede.
type Pulse struct {
	OriginNeuronID int        // ID do neurônio que disparou o pulso
	TargetNeuronID int        // ID do neurônio alvo do pulso (em uma implementação ponto a ponto)
	Position       [SpaceDimensions]float64 // Posição atual do pulso (se modelarmos a viagem)
	Strength       float64    // Força do pulso (ex: 0.3 para excitatório, -0.3 para inibitório)
	EmittedCycle   int        // Ciclo em que o pulso foi emitido

	// Para propagação baseada em área/distância, não em alvo específico inicialmente:
	CurrentRadius float64 // Raio atual de propagação do pulso desde a origem
	MaxRadius     float64 // Raio máximo que este pulso pode alcançar (pode ser global ou por tipo de neurônio)
	OriginPosition [SpaceDimensions]float64 // Posição de origem do pulso
	SourceNeuronType NeuronType // Tipo do neurônio que emitiu o pulso
}

// NewPulse cria um novo pulso.
// Para o modelo CrowNet, um pulso não tem um TargetNeuronID específico inicialmente,
// ele se propaga e afeta neurônios em seu raio de alcance.
func NewPulse(originID int, originPos [SpaceDimensions]float64, strength float64, emittedCycle int, sourceNeuronType NeuronType) *Pulse {
	// MaxRadius pode depender do tipo de neurônio ou ser uma constante global.
	// O README menciona "raio" para os tipos de neurônios, mas isso parece ser para a
	// distribuição espacial deles, não para o alcance do pulso individual.
	// A propagação do pulso é 0.6 unidades/ciclo. A distância máxima do espaço é 8 unidades.
	// Um pulso pode, teoricamente, cruzar o espaço.
	// Vamos definir um MaxRadius efetivo para fins práticos ou deixar a lógica de ciclo de vida do pulso lidar com isso.
	// O pseudocódigo do README sugere que a propagação é iterativa e baseada no raio.

	return &Pulse{
		OriginNeuronID: originID,
		OriginPosition: originPos,
		Strength:       strength,
		EmittedCycle:   emittedCycle,
		CurrentRadius:  0.0,
		MaxRadius:      8.0, // Distância máxima do espaço, conforme README
		SourceNeuronType: sourceNeuronType,
	}
}

// UpdatePropagation atualiza o raio de propagação do pulso.
// Retorna true se o pulso ainda estiver ativo/propagando.
func (p *Pulse) UpdatePropagation(pulsePropagationSpeed float64) bool {
	p.CurrentRadius += pulsePropagationSpeed
	return p.CurrentRadius <= p.MaxRadius
}

// GetEffectiveRange retorna o início e o fim da faixa de distância
// que este pulso afeta nesta iteração/ciclo de propagação.
// Baseado no pseudocódigo:
// inicio = (int8(raio/0.6)*0.6) * iteração-1
// fim = (int8(raio/0.6)*0.6) * iteração
// Isso parece um pouco estranho. Uma interpretação mais simples:
// O pulso se expande como uma casca esférica.
// Em cada ciclo, a casca avança `pulsePropagationSpeed`.
// Neurônios dentro desta "nova" região da casca são afetados.
func (p *Pulse) GetEffectiveRange(pulsePropagationSpeed float64) (rangeStart, rangeEnd float64) {
	// rangeStart é o raio do ciclo anterior
	rangeStart = p.CurrentRadius - pulsePropagationSpeed
	if rangeStart < 0 {
		rangeStart = 0
	}
	// rangeEnd é o raio atual
	rangeEnd = p.CurrentRadius
	return rangeStart, rangeEnd
}

// IsDopaminePulse verifica se o pulso é de um neurônio dopaminérgico.
// No README: "Se for dopamina soma a quantidade de dopamina do neuronio"
// Isso sugere que pulsos de neurônios dopaminérgicos são tratados de forma especial,
// não apenas como um `Strength` positivo/negativo. Eles podem aumentar o nível de dopamina
// localmente ou contribuir para o pool global de dopamina.
// Para o MVP, vamos assumir que eles têm um 'Strength' que é interpretado de forma diferente
// pelos neurônios alvo ou pela rede.
func (p *Pulse) IsDopaminePulse() bool {
	// Esta é uma simplificação. A "quantidade de dopamina do neurônio" mencionada no README
	// pode significar que o neurônio dopaminérgico libera uma quantidade variável,
	// ou que o efeito do pulso é modular a dopamina no neurônio alvo.
	// Por agora, vamos identificar o pulso.
	return p.SourceNeuronType == DopaminergicNeuron
}

// GetStrengthParaTipoAlvo retorna a força efetiva do pulso considerando o tipo do neurônio alvo.
// Por exemplo, um pulso dopaminérgico pode não afetar o potencial diretamente, mas sim
// modular a dopamina local ou o limiar do alvo.
// Para o MVP inicial, vamos manter simples: Strength afeta o potencial.
// A lógica de dopamina será tratada separadamente na atualização da rede ou do neurônio.
func (p *Pulse) GetStrength() float64 {
	// Se o pulso for de dopamina, seu 'Strength' pode ser interpretado de forma diferente.
	// O README diz "se for dopamina soma a quantidade de dopamina do neurônio".
	// Isso é ambiguo. Pode significar:
	// 1. O neurônio alvo aumenta seu próprio nível de dopamina interna (se tiver tal propriedade).
	// 2. O pulso contribui para um nível de dopamina ambiente.
	// 3. O pulso tem um efeito de potencial normal E também um efeito dopaminérgico.
	//
	// Para o MVP, vamos assumir que pulsos de neurônios dopaminérgicos têm um `Strength` que pode ser 0
	// em termos de potencial direto, e seu efeito é puramente modular (aumentar dopamina na rede).
	// Ou, eles têm um efeito excitatório normal E liberam dopamina.
	//
	// Decisão para MVP: Pulsos dopaminérgicos têm um efeito excitatório padrão (strength positivo)
	// E a rede contabilizará sua ocorrência para aumentar os níveis de dopamina (ver network.go).
	return p.Strength
}
