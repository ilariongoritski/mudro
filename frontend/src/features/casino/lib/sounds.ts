// Casino sound effects using Web Audio API (no external files needed)

type SoundType = 'spin' | 'win' | 'bigWin' | 'jackpot' | 'lose' | 'click' | 'faucet' | 'bonus';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();

function playTone(frequency: number, duration: number, type: OscillatorType = 'sine', volume = 0.1) {
  if (audioContext.state === 'suspended') {
    audioContext.resume();
  }
  
  const oscillator = audioContext.createOscillator();
  const gainNode = audioContext.createGain();
  
  oscillator.connect(gainNode);
  gainNode.connect(audioContext.destination);
  
  oscillator.type = type;
  oscillator.frequency.value = frequency;
  
  gainNode.gain.setValueAtTime(volume, audioContext.currentTime);
  gainNode.gain.exponentialRampToValueAtTime(0.001, audioContext.currentTime + duration);
  
  oscillator.start(audioContext.currentTime);
  oscillator.stop(audioContext.currentTime + duration);
}

function playChord(frequencies: number[], duration: number, type: OscillatorType = 'sine', volume = 0.08) {
  frequencies.forEach(freq => playTone(freq, duration, type, volume));
}

export function playSound(type: SoundType, soundEnabled = true) {
  if (!soundEnabled) return;
  
  try {
    switch (type) {
      case 'spin':
        // Quick ascending tone for spin start
        playTone(440, 0.1, 'sine', 0.05);
        setTimeout(() => playTone(554, 0.1, 'sine', 0.05), 50);
        setTimeout(() => playTone(659, 0.15, 'sine', 0.05), 100);
        break;
        
      case 'win':
        // Major chord - happy win sound
        playChord([523, 659, 784], 0.4, 'triangle', 0.08);
        setTimeout(() => playChord([659, 784, 988], 0.3, 'triangle', 0.06), 150);
        break;
        
      case 'bigWin':
        // More elaborate win sound
        playChord([523, 659, 784, 1047], 0.5, 'sine', 0.1);
        setTimeout(() => playChord([659, 784, 988, 1319], 0.4, 'sine', 0.08), 200);
        setTimeout(() => playChord([784, 988, 1175, 1568], 0.6, 'triangle', 0.08), 400);
        break;
        
      case 'jackpot':
        // Epic jackpot sound - ascending arpeggio
        [523, 659, 784, 1047, 1319, 1568, 2093].forEach((freq, i) => {
          setTimeout(() => playTone(freq, 0.3, 'sine', 0.1), i * 80);
        });
        setTimeout(() => playChord([1047, 1319, 1568, 2093], 1.0, 'triangle', 0.12), 600);
        break;
        
      case 'lose':
        // Descending sad tone
        playTone(330, 0.2, 'sawtooth', 0.05);
        setTimeout(() => playTone(294, 0.2, 'sawtooth', 0.05), 100);
        setTimeout(() => playTone(262, 0.3, 'sawtooth', 0.05), 200);
        break;
        
      case 'click':
        // UI click
        playTone(800, 0.05, 'square', 0.03);
        break;
        
      case 'faucet':
        // Coin drop sound
        playTone(600, 0.08, 'sine', 0.08);
        setTimeout(() => playTone(800, 0.08, 'sine', 0.06), 60);
        setTimeout(() => playTone(1000, 0.1, 'sine', 0.05), 120);
        break;
        
      case 'bonus':
        // Magic sparkle sound
        playChord([880, 1109, 1397], 0.3, 'sine', 0.08);
        setTimeout(() => playChord([1109, 1397, 1760], 0.3, 'sine', 0.06), 150);
        break;
    }
  } catch (e) {
    // Silently fail if audio not available
    console.debug('Sound playback failed:', e);
  }
}

// Initialize audio context on first user interaction
export function initAudio() {
  if (audioContext.state === 'suspended') {
    audioContext.resume();
  }
}
