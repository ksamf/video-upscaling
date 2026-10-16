export interface VideoInfoAPI {
    id: string;
    video_path: string;
    nsfw: boolean;
    language: string;
    qualities: number[];
}

export interface StatusAPI {
    video_id: string;
    status: string;
}

export interface VideoURLAPI {
    video_path: string
}

export interface DeleteVideoAPI {
    status: string;
    message: string;
}

export interface VideoSubtitlesAPI {
    subtitles: string;
}

export interface VideoDubbingAPI {
    dubbing: string;
}

export interface AllowedQualitiesAPI {
    qualities: number[]
}