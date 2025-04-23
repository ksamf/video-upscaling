import { VideoList, VideoPlayer } from "@features/video";
import { useFetchVideo } from "@features/video/hooks/useFetchVideo";
import { Spinner } from "@heroui/react";
import { NotFoundPage } from "@pages/not-found";
import { FC, useEffect } from "react";
import { useParams } from "react-router-dom";

export const PlayerPage: FC = () => {
    const {id} = useParams()
    const {video, isLoading, isError, error, fetchVideo} = useFetchVideo(id || "")

    useEffect(() => {
        fetchVideo()
    }, [])

    if (isLoading) return <Spinner />
    if (!isLoading && isError) return <p className={`text-red-600`}>{error}</p>
    if (!isLoading && !isError && video === null) return <NotFoundPage />

    return (
        <div className={`flex p-0 lg:p-4 gap-4 w-full flex-col lg:flex-row`}>
            <div className={`w-full lg:w-3/4`}>
                <VideoPlayer
                    src={video?.video_path}
                />
            </div>
            <div className={`w-full lg:w-1/4 px-4 lg:px-0`}>
                <VideoList />
            </div>
        </div>
    )
}