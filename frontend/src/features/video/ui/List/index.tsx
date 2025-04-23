import { useRecentVideos } from "@features/video/hooks/useRecentVideos";
import { Spinner } from "@heroui/react";
import React, { FC, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { VideoCard } from "../Card";

export const VideoList: FC = () => {
    const { t } = useTranslation()
    const { list, isLoading, isError, error, fetchRecentVideos } = useRecentVideos()

    useEffect(() => {
        fetchRecentVideos()
    }, [])

    return (
        <div>
            <h2>{t("video_list.title")}</h2>
            {isLoading && <Spinner />}
            {(!isLoading && isError) && (
                <p className={`text-red-600`}>{error}</p>
            )}
            {(!isLoading && !isError && list.length === 0) && (
                <p>{t("video_list.not_found")}</p>
            )}
            {(!isLoading && !isError && list.length > 0) && (
                <ul>
                    {list.map((video) => (
                        <React.Fragment key={video.video_path}>
                            <VideoCard {...video} />
                        </React.Fragment>
                    ))}
                </ul>
            )}
        </div>
    )
}