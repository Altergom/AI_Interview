import { useState } from 'react';
import { FileUpload } from '../../components/common/FileUpload';
import { Loading } from '../../components/common/Loading';
import { parseResume } from '../../services/resume';
import { useResumeStore } from '../../store/resumeStore';
import { FILE_UPLOAD } from '../../utils/constants';

const getErrorMessage = (error: unknown): string => {
  if (error && typeof error === 'object') {
    const apiError = error as { error?: { message?: string }; message?: string };
    if (apiError.error?.message) {
      return apiError.error.message;
    }
    if (apiError.message) {
      return apiError.message;
    }
  }

  return '简历解析失败，请手动填写';
};

export const PDFUpload = () => {
  const { setResume, setParsing, parseError, setParseError } = useResumeStore();
  const [uploading, setUploading] = useState(false);

  const handleFileSelect = async (file: File) => {
    setUploading(true);
    setParsing(true);
    setParseError(null);

    try {
      const result = await parseResume(file);
      setResume(result);
    } catch (error) {
      setParseError(getErrorMessage(error));
    } finally {
      setUploading(false);
      setParsing(false);
    }
  };

  return (
    <div className="space-y-4">
      <h3 className="mb-4 text-lg font-semibold text-gray-900">上传简历（可选）</h3>

      <p className="mb-4 text-sm text-gray-600">
        上传 PDF 格式简历，系统将自动解析并填充表单。如果解析失败，您也可以跳过此步骤手动填写。
      </p>

      {uploading ? (
        <Loading text="正在解析简历..." />
      ) : (
        <FileUpload
          accept={FILE_UPLOAD.ALLOWED_TYPES.join(',')}
          maxSize={FILE_UPLOAD.MAX_SIZE}
          onFileSelect={handleFileSelect}
        />
      )}

      {parseError && (
        <div className="rounded-lg border border-yellow-200 bg-yellow-50 p-3 text-sm text-yellow-800">
          {parseError}
        </div>
      )}

      <div className="text-sm text-gray-500">
        提示：您也可以跳过上传，直接点击“下一步”手动填写简历信息。
      </div>
    </div>
  );
};
