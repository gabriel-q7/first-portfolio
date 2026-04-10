import type { TerminalLine, Command } from '$lib/types/terminal';

function line(content: string, type: TerminalLine['type'] = 'output'): Omit<TerminalLine, 'id' | 'timestamp'> {
  return { content, type };
}

const COMMANDS: Record<string, Command> = {
  help: {
    name: 'help',
    description: 'Show available commands',
    handler: () => [
      line(''),
      line('  Available commands:', 'system'),
      line(''),
      line('  help      — show this help message'),
      line('  about     — about me'),
      line('  projects  — list my projects'),
      line('  contact   — get in touch'),
      line('  clear     — clear the terminal'),
      line('')
    ]
  },

  about: {
    name: 'about',
    description: 'About me',
    handler: () => [
      line(''),
      line('  ┌─────────────────────────────────────────┐', 'system'),
      line('  │  PROFILE                                │', 'system'),
      line('  └─────────────────────────────────────────┘', 'system'),
      line(''),
      line('  Name    : Gabriel'),
      line('  Role    : Software Engineer'),
      line('  Stack   : TypeScript · Go · SvelteKit · Docker'),
      line('  Focus   : Clean code, scalable systems, great UX'),
      line(''),
      line('  I build things that are fast, minimal and maintainable.'),
      line('  Currently crafting this very portfolio. 🚀'),
      line('')
    ]
  },

  projects: {
    name: 'projects',
    description: 'List my projects',
    handler: () => [
      line(''),
      line('  ┌─────────────────────────────────────────┐', 'system'),
      line('  │  PROJECTS                               │', 'system'),
      line('  └─────────────────────────────────────────┘', 'system'),
      line(''),
      line('  [01] first-portfolio'),
      line('       Hacker-style portfolio in a monorepo.'),
      line('       SvelteKit · Tailwind · TypeScript · Docker'),
      line(''),
      line('  [02] — more coming soon —'),
      line(''),
      line('  Tip: projects data will be served via API once backend is ready.', 'system'),
      line('')
    ]
  },

  contact: {
    name: 'contact',
    description: 'Get in touch',
    handler: () => [
      line(''),
      line('  ┌─────────────────────────────────────────┐', 'system'),
      line('  │  CONTACT                                │', 'system'),
      line('  └─────────────────────────────────────────┘', 'system'),
      line(''),
      line('  GitHub  : github.com/gabriel-q7'),
      line('  Email   : hello@example.com'),
      line(''),
      line('  Feel free to reach out for collaborations or just to say hi.'),
      line('')
    ]
  }
};

export function executeCommand(input: string): Omit<TerminalLine, 'id' | 'timestamp'>[] {
  const trimmed = input.trim().toLowerCase();

  if (!trimmed) return [];

  const inputLine: Omit<TerminalLine, 'id' | 'timestamp'> = {
    content: `guest@portfolio:~$ ${input}`,
    type: 'input'
  };

  if (trimmed === 'clear') {
    return [{ content: '__CLEAR__', type: 'system' }];
  }

  const command = COMMANDS[trimmed];

  if (!command) {
    return [
      inputLine,
      { content: `  command not found: ${trimmed}. Type 'help' for available commands.`, type: 'error' }
    ];
  }

  return [inputLine, ...command.handler()];
}

export { COMMANDS };
