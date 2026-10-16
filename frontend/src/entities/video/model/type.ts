export interface VideoInfo {
    id: string;
    video_path: string;
    nsfw: boolean;
    language: string;
    qualities: number[];
}

export interface Status {
    video_id: string;
    status: string;
}

export interface VideoURL {
    video_path: string
}

export interface DeleteVideo {
    status: string;
    message: string;
}

export interface VideoSubtitles {
    subtitles: string;
}

export interface VideoDubbing {
    dubbing_url: string;
}

export interface VideoDubbing {
    dubbing_url: string;
}

export interface AllowedQualities {
    qualities: number[]
}