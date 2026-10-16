import { getVideoInfo } from "@entities/video/api";
import { VideoInfo } from "@entities/video/model/type";
import { useState } from "react";

export const useFetchVideo = (id: string) => {
    const [video, setVideo] = useState<VideoInfo | null>()
    const [loading, setLoading] = useState<boolean>(false);
    const [error, setError] = useState<null | string>(null);

    const fetchVideo = async () => {
        setError(null)
        setLoading(true)

        const res = await getVideoInfo(id);

        if (typeof res !== "string") setVideo(res);
        else setError(res)

        setLoading(false)
    }

    return {
        video,
        isLoading: loading,
        isError: error !== null,
        error,
        fetchVideo
    }
}