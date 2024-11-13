declare module 'qrcodejs' {
    export default class QRCode {
        constructor(element: HTMLElement | string, options: {
            text: string;
            width?: number;
            height?: number;
            colorDark?: string;
            colorLight?: string;
            correctLevel?: number;
        });
    }
}