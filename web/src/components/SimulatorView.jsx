import React from "react";
import { Settings2, Wallet, Zap } from "lucide-react";
import { Button } from "./ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "./ui/card";
import { Input } from "./ui/input";

const SLIDER_MIN = 4;
const SLIDER_MAX = 256;
const SLIDER_STEP = 4;

const getSliderValue = (value) => {
    if (typeof value !== "number" || Number.isNaN(value)) {
        return 24;
    }

    const clamped = Math.min(SLIDER_MAX, Math.max(SLIDER_MIN, value));
    const steps = Math.round((clamped - SLIDER_MIN) / SLIDER_STEP);
    return SLIDER_MIN + steps * SLIDER_STEP;
};

const SimulatorView = ({
    newDifficulty,
    setNewDifficulty,
    handleDifficultyUpdate,
    isUpdatingDiff,
    handleChaosMintClick,
    isMinting,
    handleGenerateWalletClick,
}) => {
    return (
        <div className="space-y-4">
            <h2 className="text-xl font-semibold tracking-tight lg:text-2xl">
                Simulador de Transacoes e Rede
            </h2>

            <Card>
                <CardHeader>
                    <CardTitle>Parametros de Consenso (Proof of Work)</CardTitle>
                    <CardDescription>
                        Define a quantidade de zeros exigida no hash do bloco.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="grid gap-4 lg:grid-cols-[1fr_120px_auto] lg:items-center">
                        <input
                            type="range"
                            min={SLIDER_MIN}
                            max={SLIDER_MAX}
                            step={SLIDER_STEP}
                            value={getSliderValue(newDifficulty)}
                            onChange={(e) => setNewDifficulty(Number(e.target.value))}
                            className="h-2 w-full cursor-pointer appearance-none rounded-lg bg-secondary"
                        />
                        <Input
                            type="number"
                            min={1}
                            step={1}
                            value={newDifficulty}
                            onChange={(e) => setNewDifficulty(Number(e.target.value))}
                        />
                        <Button onClick={handleDifficultyUpdate} disabled={isUpdatingDiff}>
                            <Settings2 className="h-4 w-4" />
                            {isUpdatingDiff ? "Aplicando..." : "Aplicar Dificuldade"}
                        </Button>
                    </div>
                    <p className="text-xs text-muted-foreground">
                        Slider: de {SLIDER_MIN} ate {SLIDER_MAX} em passos de {SLIDER_STEP}. Use o campo numerico para valores customizados fora desse intervalo.
                    </p>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Controles de Simulacao</CardTitle>
                </CardHeader>
                <CardContent className="flex flex-wrap gap-2">
                    <Button onClick={handleChaosMintClick} disabled={isMinting}>
                        <Zap className="h-4 w-4" />
                        {isMinting ? "Executando..." : "Chaos Mint (10 transacoes)"}
                    </Button>
                    <Button variant="secondary" onClick={handleGenerateWalletClick}>
                        <Wallet className="h-4 w-4" />
                        Gerar Carteira Aleatoria
                    </Button>
                </CardContent>
            </Card>
        </div>
    );
};

export default SimulatorView;
