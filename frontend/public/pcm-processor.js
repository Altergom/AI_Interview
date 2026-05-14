/**
 * PCM 采集 AudioWorkletProcessor
 *
 * 将麦克风输入的 Float32 样本降采样到 16kHz / 16bit / mono PCM，
 * 分帧（每帧 4096 样本 ~= 256ms @16kHz）后通过 MessagePort 发给主线程，
 * 主线程再通过 WebSocket 以二进制帧发送给服务端。
 *
 * 使用方式（主线程）：
 *   await audioContext.audioWorklet.addModule('/pcm-processor.js');
 *   const node = new AudioWorkletNode(audioContext, 'pcm-processor', {
 *     processorOptions: { targetSampleRate: 16000 }
 *   });
 *   node.port.onmessage = (e) => ws.send(e.data); // ArrayBuffer (Int16)
 */
class PCMProcessor extends AudioWorkletProcessor {
  constructor(options) {
    super();
    // AudioContext 的采样率（通常 44100 或 48000）
    this._inputSampleRate = sampleRate; // AudioWorkletGlobalScope 全局变量
    this._targetSampleRate = (options.processorOptions || {}).targetSampleRate || 16000;
    // 降采样比
    this._ratio = this._inputSampleRate / this._targetSampleRate;

    // 累积缓冲：收够一帧后发送（目标 4096 个 16kHz 样本）
    this._frameSize = 4096;
    this._buffer = new Float32Array(this._frameSize);
    this._bufferFill = 0;

    // 降采样状态
    this._resamplePos = 0; // 当前在输入流中的浮点位置
  }

  process(inputs) {
    // inputs[0] 是第一个输入，取 channel 0（mono）
    const input = inputs[0];
    if (!input || input.length === 0) return true;
    const channel = input[0];
    if (!channel || channel.length === 0) return true;

    let inputPos = 0;
    while (inputPos < channel.length) {
      // 线性插值降采样
      const srcIndex = Math.floor(this._resamplePos);
      const frac = this._resamplePos - srcIndex;

      let sample;
      if (srcIndex + 1 < channel.length) {
        sample = channel[srcIndex] * (1 - frac) + channel[srcIndex + 1] * frac;
      } else {
        sample = channel[srcIndex] || 0;
      }

      this._buffer[this._bufferFill++] = sample;

      // 帧满则发送
      if (this._bufferFill >= this._frameSize) {
        this._flush();
      }

      this._resamplePos += this._ratio;
      // 当 resamplePos 超过当前 channel 时推进
      if (this._resamplePos >= channel.length) {
        this._resamplePos -= channel.length;
        inputPos = channel.length; // 跳出
      } else {
        inputPos = Math.floor(this._resamplePos);
      }
    }

    return true; // 保持 processor 存活
  }

  _flush() {
    // Float32 → Int16（clip 到 [-1, 1]）
    const pcm16 = new Int16Array(this._bufferFill);
    for (let i = 0; i < this._bufferFill; i++) {
      const s = Math.max(-1, Math.min(1, this._buffer[i]));
      pcm16[i] = s < 0 ? s * 0x8000 : s * 0x7fff;
    }
    // 发送 ArrayBuffer（主线程 ws.send 可直接发二进制帧）
    this.port.postMessage(pcm16.buffer, [pcm16.buffer]);
    this._bufferFill = 0;
    this._buffer = new Float32Array(this._frameSize);
  }
}

registerProcessor('pcm-processor', PCMProcessor);
