import { FC, useRef } from "react";
import { UploadInputProps } from "./type";
import DocumentUploadIcon from "@shared/assets/icons/document-upload.svg?react"

export const UploadInput: FC<UploadInputProps> = (props) => {
    const { accept, onChange, multiple = false } = props;
    const hiddenInputRef = useRef<HTMLInputElement | null>(null);

    const handleFile = (inputFiles: File) => {
        onChange({
            filepath: URL.createObjectURL(inputFiles),
        })
    }

    const handleInputFile = (e: React.ChangeEvent<HTMLInputElement>) => {
        const inputFile = e.target.files;
        if (!inputFile || inputFile?.length === 0) return;
        const targetFile = inputFile[0];
        handleFile(targetFile)
    }

    const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        const inputFile = e.dataTransfer.files;
        if (!inputFile || inputFile?.length === 0) return;
        const targetFile = inputFile[0];
        handleFile(targetFile)
    }

    const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
        e.preventDefault();
    };

    const handleButtonClick = () => {
        if (hiddenInputRef.current) {
            hiddenInputRef.current.click();
        }
    };

    return (
        <div
            className={`flex flex-col items-center justify-center py-24 border-2 border-dashed border-gray-300 
                rounded-2xl cursor-pointer hover:border-blue-500 transition-all text-center space-y-3`}
            onClick={handleButtonClick}
            onDrop={handleDrop}
            onDragOver={handleDragOver}
        >
            <DocumentUploadIcon className="w-10 h-10 text-blue-500" />
            <p className="text-sm text-gray-600">
                Drop your file,{' '}
                <button
                    type="button"
                    className="text-blue-600 underline hover:text-blue-800 focus:outline-none"
                >
                    or click to browse
                </button>
            </p>
            <input
                ref={hiddenInputRef}
                accept={accept}
                type="file"
                multiple={multiple}
                onChange={handleInputFile}
                className="hidden"
            />
        </div>
    )
}