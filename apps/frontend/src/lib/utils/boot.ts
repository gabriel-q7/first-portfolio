import type { TerminalLine } from '$lib/types/terminal';

type LineInput = Omit<TerminalLine, 'id' | 'timestamp'>;

export const BOOT_SEQUENCE: LineInput[] = [
  { content: 'Initializing system...', type: 'boot' },
  { content: 'Loading kernel modules... OK', type: 'boot' },
  { content: 'Mounting filesystems... OK', type: 'boot' },
  { content: 'Starting network services... OK', type: 'boot' },
  { content: 'Establishing secure connection... OK', type: 'boot' },
  { content: '', type: 'boot' },
  { content: '╔══════════════════════════════════════════════╗', type: 'system' },
  { content: '║          GABRIEL  //  PORTFOLIO  v1.0        ║', type: 'system' },
  { content: '╚══════════════════════════════════════════════╝', type: 'system' },
  { content: '', type: 'boot' },
  { content: "System ready. Type 'help' to see available commands.", type: 'system' },
  { content: '', type: 'boot' }
];

export async function runBootSequence(
  addLine: (line: LineInput) => void,
  onComplete: () => void,
  delay = 80
): Promise<void> {
  for (const line of BOOT_SEQUENCE) {
    await new Promise<void>((resolve) => setTimeout(resolve, delay));
    addLine(line);
  }
  onComplete();
}
