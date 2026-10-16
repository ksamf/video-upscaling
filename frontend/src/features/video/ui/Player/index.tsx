import styles from "./style.module.scss";
import MaximizeIcon from "./icons/maximize.svg?react"
import PlayIcon from "./icons/play.svg?react"
import PauseIcon from "./icons/pause.svg?react"
import BackwardIcon from "./icons/backward-15-seconds.svg?react"
import ForwardIcon from "./icons/forward-15-seconds.svg?react"
import VolumeCross from "./icons/volume-cross.svg?react"
import VolumeLow from "./icons/volume-low.svg?react"
import VolumeHigh from "./icons/volume-high.svg?react"
import SettingsIcon from "./icons/setting-3.svg?react"
import { FC, useCallback, useEffect, useRef, useState } from "react";
import { VideoPlayerProps } from "./type";
import { formatTime } from "./lib/time";
import { AnimatePresence, motion } from "framer-motion";

export const VideoPlayer: FC<VideoPlayerProps> = (props) => {
    const { src } = props;

    const videoRef = useRef<HTMLVideoElement | null>(null);

    const [isPlay, setIsPlay] = useState<boolean>(false);
    // const [isFullScreen, setIsFullScreen] = useState<boolean>(false);
    const [currentVideoTime, setCurrentVideoTime] = useState<number>(0);
    const [isVisible, setIsVisible] = useState(false);
    const [timeoutId, setTimeoutId] = useState<number | null>(null);
    const [overController, setOverController] = useState<boolean>(false);
    const [volume, setVolume] = useState<number>(50)

    const calculateCurrentProcent = useCallback(() => {
        if (!videoRef.current) return 0;
        return (videoRef.current.currentTime * 100) / videoRef.current.duration
    }, [currentVideoTime])

    const maximaize = () => {
        if (videoRef.current) {
            // setIsFullScreen(prev => !prev)
            // videoRef.current.width = !isFullScreen ? document.body.clientWidth : 100
            // videoRef.current.height = !isFullScreen ? window.innerHeight : 100
        }
    }

    const changePlay = () => {
        setIsPlay(prev => !prev);
        if (isPlay) {
            videoRef.current?.pause()
        } else {
            videoRef.current?.play()
        }
    }

    const rewind = (direct: number) => {
        if (videoRef.current) {
            videoRef.current.currentTime = videoRef.current.currentTime + direct
        }
    }

    const resetTimer = () => {
        setIsVisible(true);
        if (timeoutId) clearTimeout(timeoutId);
        // const newTimeout = setTimeout(() => setIsVisible(false), 5000);
        // setTimeoutId(newTimeout);
    };


    const handleMouseLeave = () => {
        if (timeoutId) {
            clearTimeout(timeoutId);
            setTimeoutId(null);
        }
        setIsVisible(false);
    };

    useEffect(() => {
        if (videoRef.current) {
            videoRef.current.volume = volume / 100;
        }
    }, [volume])

    useEffect(() => {
        const interval = setInterval(() => {
            if (videoRef.current) {
                setCurrentVideoTime(Math.floor(videoRef.current.currentTime))
            }
        }, 1);

        return () => {
            if (timeoutId) clearTimeout(timeoutId);
            clearInterval(interval)
        };
    }, []);

    return (
        <motion.div
            className={styles['video-player']}
            onHoverStart={() => setIsVisible(true)}
            onHoverEnd={() => setIsVisible(false)}
            onMouseLeave={handleMouseLeave}
            onMouseMove={resetTimer}
            onClick={changePlay}
        >
            <video ref={videoRef} controls={false} width="100%" height="auto">
                <source src={src} />
            </video>
            <AnimatePresence mode="wait">
                {(isVisible || overController) && (
                    <motion.div
                        className={styles['controller']}
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: 20 }}
                    >
                        <div
                            className={styles['controller-inner']}
                            onClick={e => e.stopPropagation()}
                            onMouseOver={() => setOverController(true)}
                            onMouseOut={() => setOverController(false)}
                        >
                            <div className={styles['timeline-wrapper']}>
                                <p>{formatTime(currentVideoTime)}</p>
                                <div className={styles['timeline']} onClick={ev => {
                                    var e = ev.target as HTMLDivElement;
                                    var dim = e.getBoundingClientRect();
                                    var x = ev.clientX - dim.left;
                                    if (videoRef.current) {
                                        const time = ((x * 100) / dim.width) / 100 * videoRef.current.duration
                                        videoRef.current.currentTime = time;
                                    }
                                }}>
                                    <span
                                        className={styles['timeline-circle']}
                                        style={{
                                            left: `${calculateCurrentProcent()}%`
                                        }}
                                    />
                                    <div
                                        className={styles['timeline-active-line']}
                                        style={{
                                            width: `${calculateCurrentProcent()}%`
                                        }}
                                    />
                                </div>
                                <p>{formatTime(Math.floor(videoRef.current?.duration || 0))}</p>
                            </div>

                            <div className={styles["block1"]}>
                                <div className={styles['row']}>
                                    <button onClick={() => rewind(-15)}>
                                        <BackwardIcon />
                                    </button>
                                    <button onClick={changePlay}>
                                        {isPlay ? <PauseIcon /> : <PlayIcon />}
                                    </button>
                                    <button onClick={() => rewind(15)}>
                                        <ForwardIcon />
                                    </button>
                                    <button>
                                        {volume === 0 && <VolumeCross />}
                                        {(volume < 50 && volume !== 0) && <VolumeLow />}
                                        {volume > 50 && <VolumeHigh />}
                                        <input
                                            type="range"
                                            value={volume}
                                            onChange={e => setVolume(parseInt(e.target.value))}
                                        />
                                    </button>
                                </div>
                                <div className={styles['row']}>
                                    <button>
                                        <SettingsIcon />
                                    </button>
                                    <button onClick={maximaize}>
                                        <MaximizeIcon />
                                    </button>
                                </div>
                            </div>
                        </div>
                    </motion.div>
                )}
            </AnimatePresence >
        </motion.div>
    )
}