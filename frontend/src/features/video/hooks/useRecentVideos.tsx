import { useState } from "react"
import { fetchVideos } from "@entities/video/api";
import { VideoInfo } from "@entities/video/model/type";

export const useRecentVideos = () => {
    const [list, setList] = useState<VideoInfo[]>([])
    const [loading, setLoading] = useState<boolean>(false);
    const [error, setError] = useState<null | string>(null);

    const fetchRecentVideos = async () => {
        setError(null)
        setLoading(true)

        const res = await fetchVideos();
        
        if (typeof res !== "string") setList(res);
        else setError(res)

        setLoading(false)
    }

    return {
        fetchRecentVideos,
        list,
        isLoading: loading,
        isError: error !== null,
        error
    }
}