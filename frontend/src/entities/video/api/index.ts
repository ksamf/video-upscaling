import { AllowedQualities, DeleteVideo, VideoDubbing, VideoInfo, VideoSubtitles, VideoURL } from "../model/type"
import { mapDeleteVideo, mapStatus, mapVideoAllowedQualities, mapVideoDubbing, mapVideoInfo, mapVideoSubtitles, mapVideoURL } from "../lib/mapper"
import { AllowedQualitiesAPI, DeleteVideoAPI, StatusAPI, VideoDubbingAPI, VideoInfoAPI, VideoSubtitlesAPI, VideoURLAPI } from "./type"

const baseURL = import.meta.env.VITE_BACKEND_URL + '/api/videos';
export const fetchVideos = async (): Promise<string | VideoInfo[]> => {
    try {
        const res = await fetch(`${baseURL}/`, { method: "GET" })

        if (!res.ok) {
            const errorData = await res.json();
            console.error("HTTP Error:", res.status, errorData);
            return errorData?.detail || `Ошибка загрузки: ${res.status}`;
        }

        const data = await res.json() as VideoInfoAPI[];
        return data.map(el => mapVideoInfo(el))
    } catch (err) {
        console.error("Network/Error:", err);
        return String(err)
    }
}

export const uplaodVideo = async (formData: FormData): Promise<string | StatusAPI> => {
    try {
        const res = await fetch(`${baseURL}/`, {
            method: "POST",
            body: formData,
        })

        if (!res.ok) {
            const errorData = await res.json();
            console.error("HTTP Error:", res.status, errorData);
            return errorData?.detail || `Ошибка загрузки: ${res.status}`;
        }

        const data = await res.json() as StatusAPI;
        return mapStatus(data)
    } catch (err) {
        console.error("Network/Error:", err);
        return String(err)
    }
}

export const getVideo = async (id: string): Promise<string | VideoURL> => {
    try {
        const res = await fetch(`${baseURL}/${id}`, { method: "GET" })

        if (!res.ok) {
            const errorData = await res.json();
            console.error("HTTP Error:", res.status, errorData);
            return errorData?.detail || `Ошибка загрузки: ${res.status}`;
        }

        const data = await res.json() as VideoURLAPI;
        return mapVideoURL(data)
    } catch (err) {
        console.error("Network/Error:", err);
        return String(err)
    }
}

export const deleteVideo = async (id: string): Promise<string | DeleteVideo> => {
    try {
        const res = await fetch(`${baseURL}/${id}`, { method: "DELETE" })

        if (!res.ok) {
            const errorData = await res.json();
            console.error("HTTP Error:", res.status, errorData);
            return errorData?.detail || `Ошибка загрузки: ${res.status}`;
        }

        const data = await res.json() as DeleteVideoAPI;
        return mapDeleteVideo(data)
    } catch (err) {
        console.error("Network/Error:", err);
        return String(err)
    }
}


export const getVideoStatus = async (id: string): Promise<string | Record<any, any>> => {
    try {
        const res = await fetch(`${baseURL}/${id}/status`, { method: "GET" })

        if (!res.ok) {
            const errorData = await res.json();
            console.error("HTTP Error:", res.status, errorData);
            return errorData?.detail || `Ошибка загрузки: ${res.status}`;
        }

        const data = await res.json() as Record<any, any>;
        return data
    } catch (err) {
        console.error("Network/Error:", err);
        return String(err)
    }
}

export const getVideoInfo = async (id: string): Promise<string | VideoInfo> => {
    try {
        const res = await fetch(`${baseURL}/${id}/info`, { method: "GET" })

        if (!res.ok) {
            const errorData = await res.json();
            console.error("HTTP Error:", res.status, errorData);
            return errorData?.detail || `Ошибка загрузки: ${res.status}`;
        }

        const data = await res.json() as VideoInfoAPI;
        return mapVideoInfo(data)
    } catch (err) {
        console.error("Network/Error:", err);
        return String(err)
    }
}

export const getVideoSubtitles = async (id: string, lang: string): Promise<string | VideoSubtitles> => {
    try {
        const res = await fetch(`${baseURL}/${id}/subtitles?lang=${encodeURIComponent(lang)}`, {
            method: "GET",
        })

        if (!res.ok) {
            const errorData = await res.json();
            console.error("HTTP Error:", res.status, errorData);
            return errorData?.detail || `Ошибка загрузки: ${res.status}`;
        }

        const data = await res.json() as VideoSubtitlesAPI;
        return mapVideoSubtitles(data)
    } catch (err) {
        console.error("Network/Error:", err);
        return String(err)
    }
}

export const getVideoDubbing = async (id: string, lang: string): Promise<string | VideoDubbing> => {
    try {
        const res = await fetch(`${baseURL}/${id}/dubbing?lang=${encodeURIComponent(lang)}`, {
            method: "GET",
        })

        if (!res.ok) {
            const errorData = await res.json();
            console.error("HTTP Error:", res.status, errorData);
            return errorData?.detail || `Ошибка загрузки: ${res.status}`;
        }

        const data = await res.json() as VideoDubbingAPI;
        return mapVideoDubbing(data)
    } catch (err) {
        console.error("Network/Error:", err);
        return String(err)
    }
}

export const getVideoQualities = async (id: string): Promise<string | AllowedQualities> => {
    try {
        const res = await fetch(`${baseURL}/${id}/qualities`, {
            method: "GET",
        })

        if (!res.ok) {
            const errorData = await res.json();
            console.error("HTTP Error:", res.status, errorData);
            return errorData?.detail || `Ошибка загрузки: ${res.status}`;
        }

        const data = await res.json() as AllowedQualitiesAPI;
        return mapVideoAllowedQualities(data)
    } catch (err) {
        console.error("Network/Error:", err);
        return String(err)
    }
}