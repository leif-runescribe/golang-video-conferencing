"use client";
import React from 'react';
import { signIn } from 'next-auth/react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Github } from 'lucide-react';

export default function SignIn() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-gray-900 to-gray-800 flex items-center justify-center p-4">
      <Card className="w-full max-w-md p-8 bg-gray-800 border-gray-700">
        <div className="flex flex-col items-center space-y-6">
          <h1 className="text-3xl font-bold text-white">Welcome</h1>
          <p className="text-gray-300 text-center">
            Sign in to start or join a video conference
          </p>
          <Button
            onClick={() => signIn('github', { callbackUrl: '/' })}
            className="w-full flex items-center justify-center space-x-2"
            size="lg"
          >
            <Github className="w-5 h-5" />
            <span>Sign in with GitHub</span>
          </Button>
        </div>
      </Card>
    </div>
  );
}