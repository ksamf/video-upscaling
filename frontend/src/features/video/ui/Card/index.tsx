import { FC, useEffect, useRef, useState } from "react";
import { VideoInfo } from "@entities/video/model/type";
import { Spinner } from "@heroui/react";
import PlayIcon from "@shared/assets/icons/play.svg?react"
import { formatTime } from "@features/video/lib/time";
import { Link } from "react-router-dom";

export const VideoCard: FC<VideoInfo> = (props) => {
    const { id, video_path, qualities } = props;
    const [preview, setPreview] = useState<string | null>(null);
    const [duration, setDuration] = useState<number>(0)
    const videoRef = useRef<HTMLVideoElement | null>(null);

    useEffect(() => {
        const generatePreview = async () => {
            const video = document.createElement("video");
            video.crossOrigin = "anonymous";
            video.src = video_path;
            video.currentTime = 10;
            video.muted = true;
            video.playsInline = true;

            video.addEventListener("loadeddata", () => {
                video.currentTime = 10;
            });

            video.addEventListener("loadedmetadata", () => {
                setDuration(video.duration)
            });

            video.addEventListener("seeked", () => {
                const canvas = document.createElement("canvas");
                canvas.width = video.videoWidth;
                canvas.height = video.videoHeight;
                const ctx = canvas.getContext("2d");
                if (ctx) {
                    ctx.drawImage(video, 0, 0, canvas.width, canvas.height);
                    const imageUrl = canvas.toDataURL("image/png");
                    setPreview(imageUrl);
                }
            });
        };

        generatePreview();
    }, [video_path]);

    return (
        <Link 
            to={`/player/${id}`}
            className={`relative min-w-80 min-h-60 group overflow-hidden rounded-md cursor-pointer`}
        >
            <div className={`bg-slate-200 w-full h-full
                flex items-center justify-center`}>
                {preview ? (
                    <img src={preview} alt="Preview" className={`object-cover h-full`} />
                ) : (
                    <Spinner />
                )}
                <div className={`absolute top-0 left-0 w-full h-full flex justify-center items-center
                     bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity duration-300 z-2`}>
                        <button className={`w-10 h-10 flex items-center justify-center bg-white/20 shadow-lg rounded-xl`}>
                            <PlayIcon className={`stroke-white`} />
                        </button>
                </div>
            </div>

            <div className={`absolute bottom-0 left-0 p-4 flex justify-between items-center w-full z-3`}>
                <span className={`px-1 py-1 isolate bg-white/20 shadow-lg backdrop-blur-lg  
                    text-white font-semibold rounded-xl`}>
                    {formatTime(duration)}
                </span>
                <span className={`px-4 py-2 isolate bg-white/20 shadow-lg backdrop-blur-sm   
                    text-white font-semibold rounded-xl`}>
                    {Math.max(...qualities)}Р
                </span>
            </div>

            <video
                ref={videoRef}
                src={video_path}
                controls
                className={`hidden`}
            />
        </Link>
    );
}