import { FC, useEffect } from "react"
import { VideoCard } from "../Card"
import { useRecentVideos } from "@features/video/hooks/useRecentVideos"
import { useTranslation } from "react-i18next"
import { Spinner } from "@heroui/react"

export const VideoFeed: FC = () => {
    const { t } = useTranslation()
    const { list, isLoading, isError, error, fetchRecentVideos } = useRecentVideos()

    useEffect(() => {
        fetchRecentVideos()
    }, [])

    return (
        <div>
            {isLoading && <Spinner />}
            {(!isLoading && isError) && (
                <p className={`text-red-600`}>{error}</p>
            )}
            {(!isLoading && !isError && list.length === 0) && (
                <p>{t("video_list.not_found")}</p>
            )}
            {(!isLoading && !isError && list.length > 0) && (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {/* {list.map((video) => (
                        <React.Fragment key={video.video_path}>
                            <VideoCard {...video} />
                        </React.Fragment>
                    ))} */}
                </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10].map(() => (
                    <VideoCard
                        id={'12412412'}
                        video_path="http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4"
                        nsfw={false}
                        language={'ru'}
                        qualities={[240, 360, 1080]}
                    />
                ))}
            </div>
        </div>
    )
}