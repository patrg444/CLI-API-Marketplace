import React from 'react';
import { clsx } from 'clsx';

interface SkeletonProps {
  className?: string;
  variant?: 'text' | 'circular' | 'rectangular';
  width?: string | number;
  height?: string | number;
  animation?: 'pulse' | 'wave' | 'none';
}

export const Skeleton: React.FC<SkeletonProps> = ({
  className,
  variant = 'text',
  width,
  height,
  animation = 'pulse',
}) => {
  const baseClasses = 'bg-gray-200 rounded';
  
  const animations = {
    pulse: 'animate-pulse',
    wave: 'animate-shimmer bg-gradient-to-r from-gray-200 via-gray-100 to-gray-200 bg-[length:200%_100%]',
    none: '',
  };

  const variants = {
    text: 'rounded',
    circular: 'rounded-full',
    rectangular: 'rounded-lg',
  };

  const defaultHeights = {
    text: 'h-4',
    circular: 'h-12 w-12',
    rectangular: 'h-32',
  };

  const style: React.CSSProperties = {};
  if (width) style.width = typeof width === 'number' ? `${width}px` : width;
  if (height) style.height = typeof height === 'number' ? `${height}px` : height;

  return (
    <div
      className={clsx(
        baseClasses,
        animations[animation],
        variants[variant],
        !height && defaultHeights[variant],
        className
      )}
      style={style}
    />
  );
};

interface SkeletonTextProps {
  lines?: number;
  className?: string;
}

export const SkeletonText: React.FC<SkeletonTextProps> = ({
  lines = 3,
  className,
}) => {
  return (
    <div className={clsx('space-y-2', className)}>
      {Array.from({ length: lines }).map((_, index) => (
        <Skeleton
          key={index}
          variant="text"
          width={index === lines - 1 ? '80%' : '100%'}
        />
      ))}
    </div>
  );
};

interface SkeletonCardProps {
  showImage?: boolean;
  imageHeight?: string | number;
  className?: string;
}

export const SkeletonCard: React.FC<SkeletonCardProps> = ({
  showImage = true,
  imageHeight = 200,
  className,
}) => {
  return (
    <div
      className={clsx(
        'bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden',
        className
      )}
    >
      {showImage && (
        <Skeleton variant="rectangular" height={imageHeight} className="w-full" />
      )}
      <div className="p-6">
        <Skeleton variant="text" width="60%" height={24} className="mb-2" />
        <SkeletonText lines={2} className="mb-4" />
        <div className="flex gap-2">
          <Skeleton variant="rectangular" width={80} height={32} />
          <Skeleton variant="rectangular" width={80} height={32} />
        </div>
      </div>
    </div>
  );
};

export default Skeleton;