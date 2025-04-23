
interface UploadData {
    filepath: string;
}

export interface UploadInputProps {
    accept?: string;
    multiple?: boolean;
    onChange: (data: UploadData) => void;
}