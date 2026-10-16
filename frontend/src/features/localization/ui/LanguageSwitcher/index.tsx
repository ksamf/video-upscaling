import { Dropdown, DropdownItem, DropdownMenu, DropdownTrigger } from "@heroui/react";
import { FC, useMemo } from "react";
import { useTranslation } from "react-i18next";
import LangIcon from "@shared/assets/icons/lang.svg?react"

export const LanguageSwitcher: FC = () => {
    const { t, i18n } = useTranslation();
    const langs = Array.isArray(i18n.options.supportedLngs) ?
        i18n.options.supportedLngs.filter(el => el !== 'cimode')
        : [];

    const currentLanguage = useMemo(
        () => i18n.language,
        [i18n.language],
    );

    return (
        <Dropdown>
            <DropdownTrigger className={`cursor-pointer outline-none`}>
                <LangIcon />
            </DropdownTrigger>
            <DropdownMenu
                selectionMode="single"
                disallowEmptySelection
                selectedKeys={new Set([currentLanguage])}
            >
                {langs.map(language => (
                    <DropdownItem
                        key={language}
                        onPress={() => i18n.changeLanguage(language)}
                    >
                        {t(`langs.${language}`)}
                    </DropdownItem>
                ))}
            </DropdownMenu>
        </Dropdown>
    )
}