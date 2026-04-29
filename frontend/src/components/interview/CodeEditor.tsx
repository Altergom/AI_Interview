import { useRef, useEffect } from 'react';
import * as monaco from 'monaco-editor';
import type { ProgrammingLanguage } from '../../types/interview';

interface CodeEditorProps {
  language: ProgrammingLanguage;
  value: string;
  onChange: (value: string) => void;
  onSubmit: () => void;
  height?: string;
}

const languageMap: Record<ProgrammingLanguage, string> = {
  java: 'java',
  python: 'python',
  go: 'go',
  cpp: 'cpp',
};

export const CodeEditor = ({
  language,
  value,
  onChange,
  onSubmit,
  height = '500px',
}: CodeEditorProps) => {
  const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!containerRef.current) return;

    const editor = monaco.editor.create(containerRef.current, {
      value,
      language: languageMap[language],
      theme: 'vs-dark',
      automaticLayout: true,
      fontSize: 14,
      minimap: { enabled: false },
      scrollBeyondLastLine: false,
    });

    editor.onDidChangeModelContent(() => {
      onChange(editor.getValue());
    });

    // 添加提交快捷键 Ctrl+Enter
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
      onSubmit();
    });

    editorRef.current = editor;

    return () => {
      editor.dispose();
    };
  }, []);

  // 语言切换时更新编辑器
  useEffect(() => {
    if (editorRef.current) {
      const model = editorRef.current.getModel();
      if (model) {
        monaco.editor.setModelLanguage(model, languageMap[language]);
      }
    }
  }, [language]);

  // 外部 value 变化时更新编辑器
  useEffect(() => {
    if (editorRef.current && editorRef.current.getValue() !== value) {
      editorRef.current.setValue(value);
    }
  }, [value]);

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between p-2 bg-gray-800 text-white">
        <span className="text-sm">语言: {language.toUpperCase()}</span>
        <button
          onClick={onSubmit}
          className="px-4 py-1 bg-blue-600 hover:bg-blue-700 rounded text-sm"
        >
          提交代码 (Ctrl+Enter)
        </button>
      </div>
      <div ref={containerRef} style={{ height }} />
    </div>
  );
};
