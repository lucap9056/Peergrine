import { FileInfo } from "@Src/storage/message/file";

export interface FileWrapper {
    file?: File
    info: FileInfo
}

export interface CompleteFileWrapper extends FileWrapper {
    file: File;
}

export class CompleteFileWrapper {
    public static HasFile = function (fileW: FileWrapper): CompleteFileWrapper {
        const { file } = fileW;

        if (!file) {
            throw new Error();
        }

        return Object.assign(fileW, { file });
    }
}