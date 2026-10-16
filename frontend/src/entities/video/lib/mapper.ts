import { AllowedQualities, DeleteVideo, Status, VideoDubbing, VideoInfo, VideoSubtitles, VideoURL } from "../model/type";
import { AllowedQualitiesAPI, DeleteVideoAPI, StatusAPI, VideoDubbingAPI, VideoInfoAPI, VideoSubtitlesAPI, VideoURLAPI } from "../api/type"

export const mapVideoInfo = (apiInfo: VideoInfoAPI): VideoInfo => {
    return {
        id: apiInfo.id,
        video_path: apiInfo.video_path,
        nsfw: apiInfo.nsfw,
        language: apiInfo.language,
        qualities: apiInfo.qualities
    }
}

export const mapStatus = (status: StatusAPI): Status => {
    return {
        video_id: status.video_id,
        status: status.status,
    }
}

export const mapVideoURL = (res: VideoURLAPI): VideoURL => {
    return {
        video_path: res.video_path
    }
}

export const mapDeleteVideo = (res: DeleteVideoAPI): DeleteVideo => {
    return {
        status: res.status,
        message: res.message,
    }
}

export const mapVideoSubtitles = (res: VideoSubtitlesAPI): VideoSubtitles => {
    return {
        subtitles: res.subtitles
    }
}

export const mapVideoDubbing = (res: VideoDubbingAPI): VideoDubbing => {
    return {
        dubbing_url: res.dubbing,
    }
}

export const mapVideoAllowedQualities = (res: AllowedQualitiesAPI): AllowedQualities => {
    return {
        qualities: res.qualities,
    }
}