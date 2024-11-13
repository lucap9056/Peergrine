/**
 * 定義檔案信息類
 */
export class FileInfo {
    id: string; // 檔案的唯一識別碼
    name: string; // 檔案名
    size: number; // 檔案大小（以字節為單位）

    /**
     * 構造函數，初始化 FileInfo 實例
     * @param file - 檔案對象
     * @param chunkSize - 每個分塊的大小（以字節為單位）
     */
    constructor(file: File) {
        this.id = crypto.randomUUID(); // 生成唯一識別碼
        this.name = file.name; // 檔案名
        this.size = file.size; // 檔案大小
    }
}

/**
 * 定義檔案分塊類
 */
export class FileChunk {
    id: string; // 分塊的唯一識別碼
    index: number; // 分塊的索引（從 0 開始）
    total: number;
    file_id: string; // 檔案的唯一識別碼
    data: string; // 分塊的數據

    /**
     * 構造函數，初始化 FileChunk 實例
     * @param id - 分塊的唯一識別碼
     * @param index - 分塊的索引
     * @param fileId - 檔案的唯一識別碼
     * @param data - 分塊的數據
     */
    constructor(fileId: string, index: number, total: number, data: string) {
        this.id = crypto.randomUUID(); // 設置分塊的唯一識別碼
        this.index = index; // 設置分塊的索引
        this.total = total
        this.file_id = fileId; // 設置檔案的唯一識別碼
        this.data = data; // 設置分塊的數據
    }
}

export interface StorageFile extends FileInfo {
    chunks: FileChunk[]
}

export class StorageFile {
    constructor(info: FileInfo) {
        Object.assign(this, info, { chunks: [] });
    }
}