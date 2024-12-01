"use client";
import React, {useState} from 'react';
import { Button } from '../../components/ui/button';
import { Card } from '../../components/ui/card';
import { Input } from '../../components/ui/input';
import { Label } from '../../components/ui/label';
import { useToast } from '../../components/ui/use-toast';
import { Video } from 'lucide-react';
import { useRouter } from 'next/navigation';

export default function CreateRoom() {
  const [roomName, setRoomName] = useState('');
  const router = useRouter()
  const { toast } = useToast();

  const handleCreateRoom = () => {
    if (!roomName.trim()) {
      toast({
        title: 'Error',
        description: 'Please enter a room name',
        variant: 'destructive',
      });
      return;
    }

    const roomId = `${roomName}`;
    router.push(`/room/${roomId}`);
  };

  return (
    <div className="min-h-screen bg-amber-400  flex items-center justify-center p-4">
      <Card className="w-full max-w-md p-8 bg-gray-800 border-gray-700">
        <div className="flex flex-col items-center mb-6">
          <Video className="w-12 h-12 text-blue-500 mb-4" />
          <h1 className="text-2xl font-bold text-white">Create New Room</h1>
        </div>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="roomName" className="text-white">Room Name</Label>
            <Input
              id="roomName"
              placeholder="Enter room name"
              value={roomName}
              onChange={(e) => setRoomName(e.target.value)}
              className="bg-gray-700 border-gray-600 text-white"
            />
          </div>

          <Button
            onClick={handleCreateRoom}
            className="w-full"
            size="lg"
          >
            Create Room
          </Button>
        </div>
      </Card>
    </div>
  );
}