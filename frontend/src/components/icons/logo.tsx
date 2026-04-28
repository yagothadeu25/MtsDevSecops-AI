import { cn } from '@/lib/utils';

interface LogoProps extends React.ImgHTMLAttributes<HTMLImageElement> {
    className?: string;
}

const Logo = ({ className, ...props }: LogoProps) => {
    return (
        <img
            alt="MtsDevSecops"
            className={cn('object-contain', className)}
            src="/logo.png"
            {...props}
        />
    );
};

export default Logo;
