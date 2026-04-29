interface AudioIndicatorProps {
  isRecording: boolean;
}

export const AudioIndicator = ({ isRecording }: AudioIndicatorProps) => {
  return (
    <div className="flex items-center justify-center">
      <div className={`w-16 h-16 rounded-full flex items-center justify-center ${
        isRecording ? 'bg-red-100 animate-pulse' : 'bg-gray-100'
      }`}>
        <svg
          className={`w-8 h-8 ${isRecording ? 'text-red-600' : 'text-gray-400'}`}
          fill="currentColor"
          viewBox="0 0 20 20"
        >
          <path
            fillRule="evenodd"
            d="M7 4a3 3 0 016 0v4a3 3 0 11-6 0V4zm4 10.93A7.001 7.001 0 0017 8a1 1 0 10-2 0A5 5 0 015 8a1 1 0 00-2 0 7.001 7.001 0 006 6.93V17H6a1 1 0 100 2h8a1 1 0 100-2h-3v-2.07z"
            clipRule="evenodd"
          />
        </svg>
      </div>
    </div>
  );
};
