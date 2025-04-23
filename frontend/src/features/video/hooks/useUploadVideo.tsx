import { uplaodVideo } from "@entities/video/api";
import { Status } from "@entities/video/model/type";
import { useState } from "react";

export const useUploadVideo = () => {
    const [loading, setLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);

    const upload = async (filepath: string): Promise<string | Status> => {
        setError(null)
        setLoading(true)

        const formData = new FormData();

        const blob = await fetch(filepath).then(r => r.blob())
        formData.append("video", blob)

        const res = await uplaodVideo(formData)
        if (typeof res === "string") setError(res);

        setLoading(false);
        return res;
    }

    return {
        upload,
        isLoading: loading,
        isError: error !== null,
        error
    }
}