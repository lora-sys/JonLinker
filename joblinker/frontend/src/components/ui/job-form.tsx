"use client";
import React, { useState } from "react";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { apiClient } from "@/lib/api_client";
import { useRouter } from "next/navigation";
import { X, Plus, Briefcase, MapPin, DollarSign, Type, List, Zap } from "lucide-react";

interface TagInputProps {
  label: string;
  placeholder: string;
  tags: string[];
  onTagsChange: (tags: string[]) => void;
  icon: React.ReactNode;
}

function TagInput({ label, placeholder, tags, onTagsChange, icon }: TagInputProps) {
  const [inputValue, setInputValue] = useState("");

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if ((e.key === "Enter" || e.key === ",") && inputValue.trim()) {
      e.preventDefault();
      const newTag = inputValue.trim();
      if (!tags.includes(newTag)) {
        onTagsChange([...tags, newTag]);
      }
      setInputValue("");
    } else if (e.key === "Backspace" && !inputValue && tags.length > 0) {
      onTagsChange(tags.slice(0, -1));
    }
  };

  const removeTag = (tagToRemove: string) => {
    onTagsChange(tags.filter((t) => t !== tagToRemove));
  };

  return (
    <div className="space-y-2">
      <Label className="flex items-center gap-2 text-neutral-300">
        {icon}
        {label}
      </Label>
      <div className="relative">
        <div className="flex flex-wrap gap-2 p-3 rounded-lg bg-zinc-900/80 border border-zinc-800 min-h-[48px]">
          {tags.map((tag) => (
            <span
              key={tag}
              className="flex items-center gap-1 px-3 py-1 rounded-full bg-cyan-500/10 text-cyan-400 text-sm border border-cyan-500/20"
            >
              {tag}
              <button
                type="button"
                onClick={() => removeTag(tag)}
                className="hover:text-cyan-200 transition-colors"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          ))}
          <input
            type="text"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={tags.length === 0 ? placeholder : ""}
            className="flex-1 min-w-[120px] bg-transparent text-sm text-white placeholder:text-neutral-500 outline-none"
          />
        </div>
      </div>
      <p className="text-xs text-neutral-500">Press Enter or comma to add tags</p>
    </div>
  );
}

export default function JobForm() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [location, setLocation] = useState("");
  const [jobType, setJobType] = useState("full-time");
  const [salaryMin, setSalaryMin] = useState("");
  const [salaryMax, setSalaryMax] = useState("");
  const [requirements, setRequirements] = useState<string[]>([]);
  const [niceToHave, setNiceToHave] = useState<string[]>([]);

  const jobTypes = [
    { value: "full-time", label: "Full-time" },
    { value: "part-time", label: "Part-time" },
    { value: "contract", label: "Contract" },
  ];

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError(null);

    if (!title.trim() || !description.trim() || requirements.length === 0) {
      setError("Please fill in all required fields");
      return;
    }

    setIsLoading(true);

    try {
      const salaryRange = salaryMin && salaryMax
        ? `${salaryMin}-${salaryMax}`
        : salaryMin || salaryMax || "";

      await apiClient.post("/api/jobs", {
        title,
        description,
        location,
        type: jobType,
        salary_range: salaryRange,
        structured: JSON.stringify({
          requirements,
          nice_to_have: niceToHave,
        }),
      });

      router.push("/jobs");
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : "Failed to create job";
      setError(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="w-full max-w-2xl mx-auto">
      <div className="relative">
        {/* Decorative geometric elements */}
        <div className="absolute -top-4 -left-4 w-20 h-20 border border-cyan-500/20 rotate-12 pointer-events-none" />
        <div className="absolute -bottom-4 -right-4 w-16 h-16 border border-cyan-500/10 -rotate-6 pointer-events-none" />

        <div className="relative bg-zinc-950/90 backdrop-blur-sm rounded-xl border border-zinc-800 p-8 shadow-[0_0_60px_-15px_rgba(6,182,212,0.15)]">
          {/* Header */}
          <div className="mb-8">
            <div className="flex items-center gap-3 mb-2">
              <div className="w-10 h-10 rounded-lg bg-cyan-500/10 border border-cyan-500/20 flex items-center justify-center">
                <Briefcase className="h-5 w-5 text-cyan-400" />
              </div>
              <h2 className="text-2xl font-bold text-white tracking-tight">
                Post a Job
              </h2>
            </div>
            <p className="text-neutral-400 text-sm">
              Define your role requirements and find the perfect candidate
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Title */}
            <div className="space-y-2">
              <Label htmlFor="title" className="flex items-center gap-2 text-neutral-300">
                <Type className="h-4 w-4 text-cyan-400/70" />
                Job Title <span className="text-red-400">*</span>
              </Label>
              <Input
                id="title"
                placeholder="Senior Full-Stack Engineer"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="bg-zinc-900/50 border-zinc-700/50 focus:border-cyan-500/50"
                required
              />
            </div>

            {/* Description */}
            <div className="space-y-2">
              <Label htmlFor="description" className="flex items-center gap-2 text-neutral-300">
                <List className="h-4 w-4 text-cyan-400/70" />
                Description <span className="text-red-400">*</span>
              </Label>
              <textarea
                id="description"
                placeholder="Describe the role, responsibilities, and what makes this opportunity unique..."
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={4}
                className="w-full rounded-lg bg-zinc-900/50 border border-zinc-700/50 px-4 py-3 text-sm text-white placeholder:text-neutral-500 focus:border-cyan-500/50 focus:outline-none focus:ring-2 focus:ring-cyan-500/10 transition-colors resize-none"
                required
              />
            </div>

            {/* Location + Type Row */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="location" className="flex items-center gap-2 text-neutral-300">
                  <MapPin className="h-4 w-4 text-cyan-400/70" />
                  Location
                </Label>
                <Input
                  id="location"
                  placeholder="Remote / New York, NY"
                  value={location}
                  onChange={(e) => setLocation(e.target.value)}
                  className="bg-zinc-900/50 border-zinc-700/50 focus:border-cyan-500/50"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="jobType" className="text-neutral-300">
                  Employment Type
                </Label>
                <select
                  id="jobType"
                  value={jobType}
                  onChange={(e) => setJobType(e.target.value)}
                  className="w-full h-10 rounded-lg bg-zinc-900/50 border border-zinc-700/50 px-3 text-sm text-white focus:border-cyan-500/50 focus:outline-none focus:ring-2 focus:ring-cyan-500/10 transition-colors appearance-none cursor-pointer"
                >
                  {jobTypes.map((type) => (
                    <option key={type.value} value={type.value}>
                      {type.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {/* Salary Range */}
            <div className="space-y-2">
              <Label className="flex items-center gap-2 text-neutral-300">
                <DollarSign className="h-4 w-4 text-cyan-400/70" />
                Salary Range (USD)
              </Label>
              <div className="flex items-center gap-3">
                <Input
                  type="number"
                  placeholder="80,000"
                  value={salaryMin}
                  onChange={(e) => setSalaryMin(e.target.value)}
                  className="bg-zinc-900/50 border-zinc-700/50 focus:border-cyan-500/50"
                />
                <span className="text-neutral-500 font-medium">to</span>
                <Input
                  type="number"
                  placeholder="150,000"
                  value={salaryMax}
                  onChange={(e) => setSalaryMax(e.target.value)}
                  className="bg-zinc-900/50 border-zinc-700/50 focus:border-cyan-500/50"
                />
              </div>
            </div>

            {/* Requirements Tags */}
            <TagInput
              label="Requirements"
              placeholder="React, TypeScript, 5+ years experience..."
              tags={requirements}
              onTagsChange={setRequirements}
              icon={<Zap className="h-4 w-4 text-cyan-400/70" />}
            />

            {/* Nice to Have Tags */}
            <TagInput
              label="Nice to Have"
              placeholder="GraphQL, AWS, Leadership experience..."
              tags={niceToHave}
              onTagsChange={setNiceToHave}
              icon={<Plus className="h-4 w-4 text-cyan-400/70" />}
            />

            {/* Error Message */}
            {error && (
              <div className="rounded-lg bg-red-500/10 border border-red-500/20 px-4 py-3 text-sm text-red-400">
                {error}
              </div>
            )}

            {/* Submit Button */}
            <div className="pt-4">
              <button
                type="submit"
                disabled={isLoading}
                className="group/btn relative w-full h-12 rounded-lg bg-gradient-to-r from-cyan-500 to-cyan-600 font-medium text-white shadow-[0_0_20px_-5px_rgba(6,182,212,0.3)] transition-all duration-300 hover:shadow-[0_0_30px_-5px_rgba(6,182,212,0.5)] hover:from-cyan-400 hover:to-cyan-500 disabled:opacity-50 disabled:cursor-not-allowed overflow-hidden"
              >
                <span className="relative z-10 flex items-center justify-center gap-2">
                  {isLoading ? (
                    <div className="h-4 w-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  ) : (
                    <>
                      <Briefcase className="h-4 w-4" />
                      Post Job
                    </>
                  )}
                </span>
                {/* Bottom gradient effect */}
                <span className="absolute inset-x-0 -bottom-px block h-px w-full bg-gradient-to-r from-transparent via-cyan-300 to-transparent opacity-0 transition-opacity duration-300 group-hover/btn:opacity-100" />
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}