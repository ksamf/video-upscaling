import { VideoFeed } from "@features/video"

export function HomePage() {
    return (
        <div className={`p-4 w-full`}>
            <div className={`flex justify-center p-24`}>
                <h2 className={`text-7xl font-semibold bg-gradient-to-r 
                    from-purple-500 via-pink-500 to-red-500 bg-clip-text text-transparent`}>Загруженные видео</h2>
            </div>
            <VideoFeed />
        </div>
    )
}