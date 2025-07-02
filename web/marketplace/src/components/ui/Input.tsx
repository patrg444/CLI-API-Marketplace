import React from 'react';
import { clsx } from 'clsx';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
  icon?: React.ReactNode;
  iconPosition?: 'left' | 'right';
  fullWidth?: boolean;
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  (
    {
      label,
      error,
      helperText,
      icon,
      iconPosition = 'left',
      fullWidth = true,
      className,
      ...props
    },
    ref
  ) => {
    const inputId = props.id || props.name;

    return (
      <div className={clsx(fullWidth && 'w-full')}>
        {label && (
          <label
            htmlFor={inputId}
            className="block text-sm font-medium text-gray-700 mb-1.5"
          >
            {label}
          </label>
        )}
        <div className="relative">
          {icon && iconPosition === 'left' && (
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <span className="text-gray-400 w-5 h-5">{icon}</span>
            </div>
          )}
          <input
            ref={ref}
            id={inputId}
            className={clsx(
              'w-full rounded-lg border px-4 py-2.5 text-gray-900 placeholder-gray-400',
              'transition-colors duration-200',
              'focus:outline-none focus:ring-2 focus:ring-opacity-20',
              icon && iconPosition === 'left' && 'pl-10',
              icon && iconPosition === 'right' && 'pr-10',
              error
                ? 'border-danger-500 focus:border-danger-500 focus:ring-danger-500'
                : 'border-gray-300 focus:border-primary-500 focus:ring-primary-500',
              className
            )}
            {...props}
          />
          {icon && iconPosition === 'right' && (
            <div className="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
              <span className="text-gray-400 w-5 h-5">{icon}</span>
            </div>
          )}
        </div>
        {(error || helperText) && (
          <p
            className={clsx(
              'mt-1 text-sm',
              error ? 'text-danger-600' : 'text-gray-500'
            )}
          >
            {error || helperText}
          </p>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';

interface TextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  error?: string;
  helperText?: string;
  fullWidth?: boolean;
}

export const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ label, error, helperText, fullWidth = true, className, ...props }, ref) => {
    const textareaId = props.id || props.name;

    return (
      <div className={clsx(fullWidth && 'w-full')}>
        {label && (
          <label
            htmlFor={textareaId}
            className="block text-sm font-medium text-gray-700 mb-1.5"
          >
            {label}
          </label>
        )}
        <textarea
          ref={ref}
          id={textareaId}
          className={clsx(
            'w-full rounded-lg border px-4 py-2.5 text-gray-900 placeholder-gray-400',
            'transition-colors duration-200',
            'focus:outline-none focus:ring-2 focus:ring-opacity-20',
            error
              ? 'border-danger-500 focus:border-danger-500 focus:ring-danger-500'
              : 'border-gray-300 focus:border-primary-500 focus:ring-primary-500',
            className
          )}
          {...props}
        />
        {(error || helperText) && (
          <p
            className={clsx(
              'mt-1 text-sm',
              error ? 'text-danger-600' : 'text-gray-500'
            )}
          >
            {error || helperText}
          </p>
        )}
      </div>
    );
  }
);

Textarea.displayName = 'Textarea';

export default Input;