import { AppRouter } from "./providers/AppRouter"
import "@shared/config/i18next"
import { HeroUIProvider } from "@heroui/system";
import "./styles/index.css"

export function App() {
    return (
        <HeroUIProvider>
            <AppRouter />
        </HeroUIProvider>
    )
}