import { BrowserRouter, Route, Routes } from "react-router-dom";
import { DefaultLayout } from "../layouts/DefaultLayout/DefaultLayout";
import { HomePage, NotFoundPage, PlayerPage } from "@pages";


export function AppRouter() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<DefaultLayout />}>
                    <Route index element={<HomePage />} />
                    <Route path="player/:id" element={<PlayerPage />} />
                    <Route path="*" element={<NotFoundPage />} />
                </Route>
            </Routes>
        </BrowserRouter>
    )
}