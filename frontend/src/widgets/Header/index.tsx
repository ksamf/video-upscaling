import { FC } from "react";
import { LanguageSwitcher } from "@features/localization/ui/LanguageSwitcher";
import { AddVideoModal } from "@features/video";
import { Link } from "react-router-dom";

export const Header: FC = () => {
    return (
        <div className={`flex justify-between p-4 items-center`}>
            <Link to="/">
                <h2 className={`font-semibold text-lg`}>FlowUp</h2>
            </Link>
            <div className={`flex items-center gap-4`}>
                <LanguageSwitcher />
                <AddVideoModal />
            </div>
        </div>
    )
}