import React from 'react';
import { Header } from './Header';
import { Footer } from './Footer';

interface ContainerProps {
  children: React.ReactNode;
  showHeader?: boolean;
  showFooter?: boolean;
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
  className?: string;
}

export const Container: React.FC<ContainerProps> = ({
  children,
  showHeader = true,
  showFooter = true,
  maxWidth = 'xl',
  className = '',
}) => {
  const maxWidthStyles = {
    sm: 'max-w-2xl',
    md: 'max-w-4xl',
    lg: 'max-w-6xl',
    xl: 'max-w-7xl',
    full: 'max-w-full',
  };

  return (
    <div className="min-h-screen flex flex-col bg-gray-50">
      {showHeader && <Header />}
      <main className={`flex-1 ${maxWidthStyles[maxWidth]} w-full mx-auto px-4 sm:px-6 lg:px-8 py-8 ${className}`}>
        {children}
      </main>
      {showFooter && <Footer />}
    </div>
  );
};
