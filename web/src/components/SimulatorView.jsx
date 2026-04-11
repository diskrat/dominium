import React from "react";
import { Settings2, Wallet, Zap, ChevronUp, ChevronDown } from "lucide-react";
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
    chaosMintCount,
    setChaosMintCount,
}) => {
    const clamp = (v, min, max) => Math.max(min, Math.min(max, v));

    const incrementDifficulty = () => {
        const v = typeof newDifficulty === "number" ? newDifficulty : Number(newDifficulty) || SLIDER_MIN;
        setNewDifficulty(clamp(v + SLIDER_STEP, SLIDER_MIN, SLIDER_MAX));
    };

    const decrementDifficulty = () => {
        const v = typeof newDifficulty === "number" ? newDifficulty : Number(newDifficulty) || SLIDER_MIN;
        setNewDifficulty(clamp(v - SLIDER_STEP, SLIDER_MIN, SLIDER_MAX));
    };

    const incrementChaos = () => {
        const v = Number(chaosMintCount) || 1;
        setChaosMintCount(Math.min(500, v + 1));
    };

    const decrementChaos = () => {
        const v = Number(chaosMintCount) || 1;
        setChaosMintCount(Math.max(1, v - 1));
    };
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
                    <div className="grid gap-4 lg:grid-cols-[1fr_auto] lg:items-center">
                        <input
                            type="range"
                            min={SLIDER_MIN}
                            max={SLIDER_MAX}
                            step={SLIDER_STEP}
                            value={getSliderValue(newDifficulty)}
                            onChange={(e) => setNewDifficulty(Number(e.target.value))}
                            className="h-2 w-full cursor-pointer appearance-none rounded-lg bg-secondary"
                        />
                        <div className="flex items-center overflow-hidden rounded-md border border-border bg-background/70 shadow-sm divide-x divide-border">
                            <Input
                                type="number"
                                min={1}
                                step={1}
                                value={newDifficulty}
                                onChange={(e) => setNewDifficulty(Number(e.target.value))}
                                className="w-28 text-base text-center rounded-none border-0 hide-native-spinner"
                            />
                            <div className="flex flex-col items-center justify-center px-1 divide-y divide-border">
                                <button
                                    type="button"
                                    onClick={incrementDifficulty}
                                    disabled={isUpdatingDiff}
                                    className="h-5 w-6 flex items-center justify-center text-foreground hover:bg-secondary/10 disabled:opacity-50 focus:outline-none"
                                    aria-label="Increment difficulty"
                                >
                                    <ChevronUp className="h-3 w-3" />
                                </button>
                                <button
                                    type="button"
                                    onClick={decrementDifficulty}
                                    disabled={isUpdatingDiff}
                                    className="h-5 w-6 flex items-center justify-center text-foreground hover:bg-secondary/10 disabled:opacity-50 focus:outline-none"
                                    aria-label="Decrement difficulty"
                                >
                                    <ChevronDown className="h-3 w-3" />
                                </button>
                            </div>
                            <Button onClick={handleDifficultyUpdate} disabled={isUpdatingDiff} className="rounded-none px-3">
                                <Settings2 className="h-4 w-4" />
                                {isUpdatingDiff ? "Aplicando..." : "Aplicar Dificuldade"}
                            </Button>
                        </div>
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
                <CardContent className="flex flex-wrap gap-2 items-center">
                    <div className="flex items-center overflow-hidden rounded-md border border-border bg-background/70 shadow-sm divide-x divide-border">
                        <Button
                            onClick={() => handleChaosMintClick(chaosMintCount)}
                            disabled={isMinting}
                            className="rounded-none px-3"
                        >
                            <Zap className="h-4 w-4" />
                            {isMinting ? "Executando..." : "Chaos Mint"}
                        </Button>
                        <Input
                            type="number"
                            min={1}
                            max={500}
                            value={chaosMintCount}
                            onChange={(e) => setChaosMintCount(Number(e.target.value) || 1)}
                            className="w-28 text-base text-center rounded-none border-0 hide-native-spinner"
                            aria-label="Quantidade de transacoes"
                        />
                        <div className="flex flex-col items-center justify-center px-1 divide-y divide-border">
                            <button
                                type="button"
                                onClick={incrementChaos}
                                disabled={isMinting}
                                className="h-5 w-6 flex items-center justify-center text-foreground hover:bg-secondary/10 disabled:opacity-50 focus:outline-none"
                                aria-label="Increment chaos mint count"
                            >
                                <ChevronUp className="h-3 w-3" />
                            </button>
                            <button
                                type="button"
                                onClick={decrementChaos}
                                disabled={isMinting}
                                className="h-5 w-6 flex items-center justify-center text-foreground hover:bg-secondary/10 disabled:opacity-50 focus:outline-none"
                                aria-label="Decrement chaos mint count"
                            >
                                <ChevronDown className="h-3 w-3" />
                            </button>
                        </div>
                    </div>
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
